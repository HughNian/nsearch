package indexer

import (
	"log"
	"sync"
	"nsearch/constant"
	"nsearch/utils"
	"strings"
	"bytes"
	"encoding/gob"
)

//索引器工作协程
var (
	indexer   *IndexWorker
	once      sync.Once
)

type IndexWorker struct {
	inited   bool

	index    *Index
	Request  chan *IndexerRequest
	Srequest chan *SearchRequest
	Srespone chan *SearchRespone
}

type IndexerRequest struct {
	DocId      int
	DocType    int

	Content    string
	WordsNum   float32
	Words      []string

	Delete     bool  //是否删除索引
}

type SearchRequest struct {
	QueryId        uint64            //查询id
	Query          string            //查询短语
	WordsNum       float32           //查询分词数量
	Words          []string          //查询分词
	WordsRecords   map[string][]byte //持久层中的记录
	Mode           int               //搜索模式，1-模糊搜索，只要有一个词匹配就可以，默认为模糊搜索，2-精确搜索，只要有一个词不配就不没结果
	Page           int
	Limit          int
}

type SearchRespone struct {
	QueryId     uint64              //查询id
	Query       string              //查询短语
	WordsNum    float32             //查询分词数量
	Words       []string            //查询分词
	WordDocNum  map[string]float32  //查询分词出现该词的文档数
	InterDocs   []*Document         //查询结果的文档
}

func NewIndexWorker() *IndexWorker {
	if indexer != nil && indexer.inited == true {
		log.Fatal("单实例不能再次初始化")
	}

	if indexer == nil {
		once.Do(func() {
			indexer = &IndexWorker {
				inited : true,
				index  : NewIndex(),
				Request  : make(chan *IndexerRequest, constant.CHAN_SIZE),
				Srequest : make(chan *SearchRequest, constant.CHAN_SIZE),
				Srespone : make(chan *SearchRespone, constant.CHAN_SIZE),
			}
		})
	}

	return indexer
}

func (ier *IndexWorker) Index() *Index {
	return ier.index
}

func (ier *IndexWorker) DoIndex() {
	for true {
		request := <- ier.Request

		if request.DocId != 0 && request.DocType != 0 && !request.Delete { //添加、更新索引记录
			document := ier.index.CreateDocument(
				request.DocId,
				request.DocType,
				request.WordsNum,
				request.Words,
				request.Content,
			)
			for _, word := range request.Words {
				wsi := utils.GetWordsInfo(word)
				if wsi != nil {
					//fmt.Println("add index word:", wsi[0])
					//fmt.Println("add index content:", request.Content)
					ier.index.Add(strings.TrimSpace(wsi[0]), document)
				}
			}
		} else if request.DocId != 0 && request.DocType != 0 && request.Delete { //删除索引记录
			for _, word := range request.Words {
				wsi := utils.GetWordsInfo(word)
				if wsi != nil {
					ier.index.delRecordByIdType(request.DocId, request.DocType, strings.TrimSpace(wsi[0]))
				}
			}
		}
	}
}

func (ier *IndexWorker) FindIndex() {
	for {
		request := <- ier.Srequest

		if len(request.Query) > 0 && request.WordsNum > 0 {
			allDocuments := make([][]*Document, len(request.Words))
			allWordDocuments := make(map[string][]*Document, len(request.Words))
			wordDocNum   := make(map[string]float32, len(request.Words))

			for k, word := range request.Words {
				//index索引中的记录
				keyId := utils.GetKeysId(word)
				documents1 := ier.index.Find(keyId)
				if documents1 != nil {
					allDocuments[k] = documents1
					allWordDocuments[word] = documents1
					wordDocNum[word] = float32(len(documents1))
				} else {
					wordDocNum[word] = float32(0)
					allDocuments[k] = nil
					allWordDocuments[word] = nil
				}

				//db持久层中的记录
				var documents2 []*Document
				if request.WordsRecords != nil {
					if record, exist := request.WordsRecords[word]; exist {
						buf := bytes.NewReader(record)
						dec := gob.NewDecoder(buf)
						err := dec.Decode(&documents2)
						if err == nil {
							//fmt.Println("key word:", word)
							//fmt.Println("data len", len(documents2))
							//for _, doc := range documents2 {
							//	fmt.Println("docid:", doc.DocId, "doctype:", doc.DocType, "content:", doc.Content)
							//}
						}
					}
					if documents2 != nil {
						if documents1 != nil && allDocuments[k] != nil && wordDocNum[word] > 0 {
							allDocuments[k]  = append(allDocuments[k], documents2...)
							allWordDocuments[word] = append(allWordDocuments[word], documents2...)
							wordDocNum[word] = float32(len(documents1)) + float32(len(documents2))
						} else {
							allDocuments[k]  = documents2
							wordDocNum[word] = float32(len(documents2))
							allWordDocuments[word] = documents2
						}
					}
				}

				//搜索模式处理
				if request.Mode != 1 && len(documents1) == 0 && len(documents2) == 0 {
					allDocuments = nil
					break
				}
			}

			realAllDocs := make([][]*Document, 0)
			for _, docs := range allDocuments {
				if len(docs) != 0 {
					realAllDocs = append(realAllDocs, docs)
				}
			}

			var pData []*Document
			if realAllDocs != nil {
				interDocs := getDocIntersect(realAllDocs, request.Mode)
				pData = pageData(interDocs, request.Page, request.Limit) //分页
			}

			ier.Srespone <- &SearchRespone {
				QueryId    : request.QueryId,
				Query      : request.Query,
				WordsNum   : request.WordsNum,
				Words      : request.Words,
				WordDocNum : wordDocNum,
				InterDocs  : pData,
			}
		}
	}
}

func (ier *IndexWorker) UpdateStorageStatus(keyword string, status bool) {
	keyId := utils.GetKeysId(keyword)
	ier.index.UpateStorageStatus(keyId, status)
}

func (ier *IndexWorker) DelStorageIndex(keyword string) {
	go func() {
		keyId := utils.GetKeysId(keyword)
		ier.index.Del(keyId)
	}()
}

//获取文档的交集
func getDocIntersect(all [][]*Document, mode int) []*Document {
	l := len(all)
	if l == 0 {
		return nil
	}
	if l == 1 {
		return all[0]
	}

	var doc []*Document
	for n := 0; n < l; n++ {
		if n == 1 {
			doc = getIntersect(all[0], all[1])

			if len(doc) == 0 && l != 2 {
				if mode != 1 {
					return nil
				} else {
					doc = all[1]
				}
			}
		} else if n > 1 {
			doc2 := doc
			doc = getIntersect(doc, all[n])
			if len(doc) == 0 {
				if mode != 1 {
					return nil
				} else {
					doc = doc2
				}
			}
		}
	}

	return doc
}

//获取两个文档的交集
func getIntersect(one, two []*Document) (inter []*Document) {
	inter = make([]*Document, 0)

	for i := 0; i < len(one); i++ {
		d1 := one[i]
		//fmt.Println("d1 docid:", d1.DocId, "d1 doctype:", d1.DocType, "d1 content:", d1.Content)
		if contains(two, d1) {
			inter = append(inter, d1)
		}
	}

	return inter
}

func contains(a []*Document, b *Document) bool {
	if len(a) == 0 {
		return false
	}

	for i := 0; i < len(a); i++ {
		//fmt.Println("d2 docid:", a[i].DocId, "d2 doctype:", a[i].DocType, "d2 content:", a[i].Content)
		if a[i].DocId == b.DocId && a[i].DocType == b.DocType {
			return true
		}
	}

	return false
}

func pageData (data []*Document, page, limit int) []*Document {
	if data == nil || len(data) == 0 {
		return nil
	}

	//default
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}

	var total, start, end int
	length := len(data)
	total = length / limit
	if page > total {
		start = (page - 1) * limit
		if start > length {
			return nil
		}
		end   = length
	} else {
		start = (page - 1) * limit
		end   = start + limit
	}

	return data[start:end]
}