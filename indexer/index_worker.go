package indexer

import (
	"log"
	"sync"
	"nsearch/constant"
	"nsearch/utils"
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

	Update     bool
}

type SearchRequest struct {
	QueryId    int         //查询id
	Query      string      //查询短语
	WordsNum   float32     //查询分词数量
	Words      []string    //查询分词
}

type SearchRespone struct {
	QueryId     int                 //查询id
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

func (ier *IndexWorker) AddIndex() {
	for true {
		request := <- ier.Request

		if request.DocId != 0 && request.DocType != 0 {
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
					ier.index.Add(wsi[0], document)
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
				keyId := utils.GetKeysId(word)
				documents := ier.index.Find(keyId)
				if documents != nil {
					allDocuments[k] = documents
					allWordDocuments[word] = documents
					wordDocNum[word] = float32(len(documents))
				} else {
					wordDocNum[word] = float32(0)
					allWordDocuments[word] = nil
				}
			}
			//for w, docs := range allWordDocuments {
			//	fmt.Println("word:", w)
			//	for _, doc := range docs {
			//		fmt.Println("content:", doc.Content)
			//	}
			//	fmt.Printf("\n\n")
			//}

			interDocs := getDocIntersect(allDocuments)
			//for k, idoc := range interDocs {
			//	fmt.Println("k--:", k, "idoc--:", idoc.Content)
			//}
			ier.Srespone <- &SearchRespone {
				QueryId    : request.QueryId,
				Query      : request.Query,
				WordsNum   : request.WordsNum,
				Words      : request.Words,
				WordDocNum : wordDocNum,
				InterDocs  : interDocs,
			}
		}
	}
}

//获取文档的交集
func getDocIntersect(all [][]*Document) []*Document {
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
				doc = all[1]
			}
		} else {
			doc2 := doc
			doc = getIntersect(doc, all[n])
			if len(doc) == 0 {
				doc = doc2
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
		if a[i].DocId == b.DocId && a[i].DocType == b.DocType {
			return true
		}
	}

	return false
}