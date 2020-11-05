package indexer

import (
	"github.com/google/btree"
	"sync"
	"nsearch/utils"
	"strconv"
)

//索引器
type Index struct {
	btLock struct {
		sync.Mutex
		bt   *btree.BTree
	}

	DocAllNum       uint64      //文档总数
	DocAllWordsNum  float32     //文档总的关键词数
}

type DocumentRecords struct {
	keyId      int
	keyword    string
	docNum     int              //该词的文档总数量
	documents  []*Document
	storge     bool
}

type Document struct {
	DocId       int
	DocType     int
	ContentLen  int
	Content     string

	Distance    float64             //该文档内容词频的总距离

	WordsNum    float32             //该文档总词数
	WordsLen    int                 //该文档词的总长度
	Words       []string            //该文档所有分词
	WordsTime   map[string]int      //该文档所有所有分词在该文档中出现的次数
	WordsFreq   map[string]float64  //该文档所有所有分词在该文档中的词频

	LabelsNum   int
	Labels      []string
}

func NewIndex() *Index {
	index := &Index {
		DocAllNum : uint64(0),
		DocAllWordsNum : float32(0),
	}

	index.btLock.bt = btree.New(2)

	return index
}

func (records *DocumentRecords) Less(item btree.Item) bool {
	return records.keyId < (item.(*DocumentRecords)).keyId
}

func (i *Index) Add(word string, doc *Document) {
	i.btLock.Lock()
	i.addRecord(word, doc)
	i.btLock.Unlock()
}

func (i *Index) Find(keyId int) []*Document {
	i.btLock.Lock()
	defer i.btLock.Unlock()

	item := i.getRecord(keyId)
	if item != nil {
		record := item.(*DocumentRecords)
		if !record.storge {
			return record.documents
		}
	}

	return nil
}

func (i *Index) Del(keyId int) *DocumentRecords {
	i.btLock.Lock()
	defer i.btLock.Unlock()

	item := i.delRecord(keyId)
	if item != nil {
		return item.(*DocumentRecords)
	}

	return nil
}

func (i *Index) UpateStorageStatus(keyId int, status bool) {
	i.btLock.Lock()
	defer i.btLock.Unlock()

	item := i.getRecord(keyId)
	if item != nil {
		docRecord := item.(*DocumentRecords)
		docRecord.storge = status
	}
}

func (i *Index) addRecord(word string, doc *Document) {
	var documents []*Document

	keyId := utils.GetKeysId(word)
	item := i.getRecord(keyId)
	if item != nil {
		record := item.(*DocumentRecords)
		documents = record.documents
	} else {
		documents = make([]*Document, 0)
	}
	if len(documents) > 0 {
		for _, d := range documents {
			if d.DocId == doc.DocId &&
			   d.DocType == doc.DocType {
				return
			}
		}

		documents = append(documents, doc)
	} else {
		documents = append(documents, doc)
	}

	i.btLock.bt.ReplaceOrInsert(&DocumentRecords{
		keyId     : keyId,
		keyword   : word,
		docNum    : len(documents),
		documents : documents,
		storge    : false,
	})

	/*keyId2 := utils.GetKeysId("明朝")
	fmt.Println("keyId2:", keyId2)
	item2 := i.getRecord(keyId2)
	if item2 != nil {
		record := item2.(*DocumentRecords)
		fmt.Println("doc len:", len(record.documents))

		for _, d := range record.documents {
			fmt.Println("doc content:", d.Content)
		}
	}*/
}

func (i *Index) delRecord(keyId int) btree.Item {
	item := i.btLock.bt.Delete(&DocumentRecords{keyId:keyId})

	return item
}

func (i *Index) getRecord(keyId int) btree.Item {
	item := i.btLock.bt.Get(&DocumentRecords{keyId:keyId})

	return item
}

func (i *Index) ForeachRecord() (outRecords []map[string][]*Document) {
	outRecords = make([]map[string][]*Document, 0)
	i.btLock.bt.Descend(func(i btree.Item) bool {
		outRecord := make(map[string][]*Document)
		v := i.(*DocumentRecords)
		outRecord[v.keyword] = v.documents
		outRecords = append(outRecords, outRecord)

		return true
	})

	return outRecords
}

func (i *Index) CreateDocument(docId, docType int, wordsNum float32, words []string, content string) *Document {
	var (
		distance = 0.
		wLen = 0
		realWords []string
	)

	i.DocAllNum++
	i.DocAllWordsNum += wordsNum

	for _, w := range words {
		wsi := utils.GetWordsInfo(w)
		if wsi != nil {
			wd, _ := strconv.ParseFloat(wsi[1], 64)
			distance += wd

			realWords = append(realWords, wsi[0])
			wLen += len([]rune(wsi[0]))
		}
	}

	//计算所有分词在该文档中出现的次数
	wtime := make(map[string]int, len(realWords))
	for _, w := range realWords {
		wtime[w]++
	}

	//计算所有分词在该文档中的词频
	wfreq := make(map[string]float64, len(realWords))
	for _, w := range realWords {
		wfreq[w] = float64(float64(wtime[w]) / float64(wordsNum))
	}

	return &Document {
		DocId      : docId,
		DocType    : docType,
		ContentLen : len([]rune(content)),
		Content    : content,
		Distance   : distance,
		WordsLen   : wLen,
		WordsNum   : wordsNum,
		Words      : realWords,
		WordsTime  : wtime,
		WordsFreq  : wfreq,
	}
}

func GetDocByTypeId(docs []*Document, docType int, docId int) *Document {
	if len(docs) == 0 {
		return nil
	}

	for _, doc := range docs {
		if doc.DocType == docType && doc.DocId == docId {
			return doc
		}
	}

	return nil
}

func UpdateDocByTypeId(docs []*Document, ndoc *Document) {
	if len(docs) == 0 || ndoc == nil {
		return
	}

	for _, doc := range docs {
		if doc.DocType == ndoc.DocType && doc.DocId == ndoc.DocId {
			if doc.Content != ndoc.Content {
				doc.Content = ndoc.Content
			}
		}
	}
}