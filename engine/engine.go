package engine

import (
	"nsearch/indexer"
	"nsearch/parter"
	"nsearch/constant"
	"nsearch/ranker"
	"nsearch/utils"
	"nsearch/include"
	"nsearch/storage"
	"sync"
	"log"
	"strings"
	"fmt"
	"bytes"
	"encoding/gob"
)

var (
	eng    *Engine
	once   sync.Once
)

type Engine struct {
	inited     bool

	stopWords  *StopWords

	iworker    *indexer.IndexWorker
	pworker    *parter.ParterWorker
	rworker    *ranker.RankerWorker
	sworker    *storage.StorageWorker
}

type QueryRequest struct {
	query      string
	retCall    RetCall
}

type RetCall func(result []*include.RetDocument) ([]byte, error)

func NewEngine() *Engine {
	if eng != nil && eng.inited == true {
		log.Fatal("单实例不能再次初始化")
	}

	if eng == nil {
		once.Do(func() {
			eng = &Engine {
				inited    : true,
				stopWords : NewStopWords("./data/stop_words.txt"),
				iworker   : indexer.NewIndexWorker(),
				pworker   : parter.NewParterWorker(),
				rworker   : ranker.NewRankerWorker(),
				sworker   : storage.NewStorageWorker(constant.DB_ENGINE),
			}
		})

		//运行分词器
		go eng.pworker.DoParter()
		//运行搜索器，添加更新删除索引
		go eng.iworker.DoIndex()
		//允许搜索器，搜索
		go eng.iworker.FindIndex()
		//运行排序器
		go eng.rworker.DocRank()
		//运行存储器
		go eng.sworker.DoStorage()
	}

	return eng
}

//把内容加入索引
func (e *Engine) IndexDoc(docId, docType int, content string) {
	if !e.inited {
		fmt.Println("搜索引擎必须初始化")
		return
	}

	if len(content) == 0 {
		fmt.Println("缺少索引内容")
		return
	}

	//分词请求
	e.pworker.Request <- &parter.PaterRequest {
		ParterMode : constant.PART_MODE_TWO,
		ParterType : constant.PARTER_TYPE_ONE,
		ParterTag  : "2", //这里不能为"0"
		DocId   : docId,
		DocType : docType,
		Content : content,
		Result  : func(ret []byte) (interface{}, error) {
			if len(ret) > 0 {
				words := strings.Split(string(ret), "|")
				wordsNum := len(words)
				if wordsNum > 0 {
					var useWords []string
					for _, word := range words {
						wsi := utils.GetWordsInfo(word)
						if wsi != nil {
							if e.stopWords.StopWordsExist(wsi[0]) {
								wordsNum--
							} else {
								useWords = append(useWords, word)
							}
						}
					}

					if len(useWords) > 0 {
						e.iworker.Request <- &indexer.IndexerRequest {
							DocId    : docId,
							DocType  : docType,
							Content  : content,
							WordsNum : float32(wordsNum),
							Words    : useWords,
							Delete   : false,
						}
					}
				}
			}

			return nil, nil
		},
	}
}

//删除内容的索引
func (e *Engine) DelIndexDoc(docId, docType int, content string) {
	if !e.inited {
		fmt.Println("搜索引擎必须初始化")
		return
	}

	if len(content) == 0 {
		fmt.Println("缺少索引内容")
		return
	}

	//分词请求
	e.pworker.Request <- &parter.PaterRequest {
		ParterMode : constant.PART_MODE_TWO,
		ParterType : constant.PARTER_TYPE_ONE,
		ParterTag  : "2",
		DocId   : docId,
		DocType : docType,
		Content : content,
		Result  : func(ret []byte) (interface{}, error) {
			if len(ret) > 0 {
				words := strings.Split(string(ret), "|")
				wordsNum := len(words)
				if wordsNum > 0 {
					var useWords []string
					for _, word := range words {
						wsi := utils.GetWordsInfo(word)
						if wsi != nil {
							if e.stopWords.StopWordsExist(wsi[0]) {
								wordsNum--
							} else {
								useWords = append(useWords, word)
							}
						}
					}

					if len(useWords) > 0 {
						//删除btree索引记录
						e.iworker.Request <- &indexer.IndexerRequest {
							DocId    : docId,
							DocType  : docType,
							Content  : content,
							WordsNum : float32(wordsNum),
							Words    : useWords,
							Delete   : true,
						}

						//删除持久层索引记录
						e.sworker.Srequest <- &storage.StorageRequest {
							DocId    : docId,
							DocType  : docType,
							Content  : content,
							WordsNum : float32(wordsNum),
							Words    : useWords,
							Delete   : true,
						}
					}
				}
			}

			return nil, nil
		},
	}
}

//刷新索引，把新建的索引加入到持久层
func (e *Engine) FlushIndex() {
	if !e.inited {
		log.Fatal("搜索引擎必须初始化")
	}

	go func() {
		indexDocs := e.iworker.Index().ForeachRecord()
		indexDocsLen := len(indexDocs)
		if indexDocsLen <= 0 {
			return
		} else {
			records := make(map[string][]byte, indexDocsLen)
			for _, keydoc := range indexDocs {
				for key, doc := range keydoc {
					//更新记录持久存储的状态为true
					e.iworker.UpdateStorageStatus(key, true)

					//从btree中删除持久存储状态为true的记录
					e.iworker.DelStorageIndex(key)

					//gob encode
					var value bytes.Buffer
					enc := gob.NewEncoder(&value)
					err := enc.Encode(doc)
					if err == nil {
						records[key] = value.Bytes()
					}
				}
			}

			e.sworker.Record <- records
		}
	}()
}

//搜索
func (e *Engine) NSearch(query string, mode, page, limit int, retCall RetCall) {
	if !e.inited {
		log.Fatal("搜索引擎必须初始化")
		return
	}

	if len(query) == 0 {
		log.Fatalln("缺少搜索内容")
		return
	}

	//搜索查询请求
	queryReq := &QueryRequest {
		query   : query,
		retCall : retCall,
	}

	//分词请求
	e.pworker.Request <- &parter.PaterRequest {
		ParterMode : constant.PART_MODE_TWO,
		ParterType : constant.PARTER_TYPE_TWO,
		ParterTag  : "0",
		Content    : queryReq.query,
		QResult    : func(ret []byte) (interface{}, error) {
			fmt.Println("search part words", string(ret))
			if len(ret) > 0 {
				words := strings.Split(string(ret), "|")
				wordsNum := len(words)
				if wordsNum > 0 {
					var (
						qid             uint64
						useWords        []string
					)
					wordsRecords := make(map[string][]byte)
					for _, word := range words {
						if len(word) != 0 && !e.stopWords.StopWordsExist(word) {
							useWords = append(useWords, word)
							//查询持久层记录
							record, err := e.sworker.Istorage.GetData([]byte(word))
							if err == nil {
								wordsRecords[word] = record
							}
						}
					}

					qid = utils.GetKeysId(useWords...)
					e.iworker.Srequest <- &indexer.SearchRequest {
						QueryId      : qid,
						Query        : query,
						WordsNum     : float32(wordsNum),
						Words        : useWords,
						WordsRecords : wordsRecords,
						Mode         : mode,
						Page         : page,
						Limit        : limit,
					}

					sresponse := <- e.iworker.Srespone
					if sresponse.InterDocs != nil && len(sresponse.InterDocs) > 0 {
						if sresponse != nil && qid == sresponse.QueryId {
							e.rworker.Rank().DocAllNum = e.iworker.Index().DocAllNum
							e.rworker.Rank().DocAllWordsNum = e.iworker.Index().DocAllWordsNum

							var rankDocs []*ranker.RankDocument
							if sresponse.InterDocs != nil && len(sresponse.InterDocs) > 0 {
								for _, doc := range sresponse.InterDocs {
									rankDocs = append(rankDocs, e.rworker.NewRankDocument(
										doc.DocId,
										doc.DocType,
										doc.ContentLen,
										doc.Content,
										doc.WordsNum,
										doc.WordsLen,
										doc.Words,
										doc.WordsTime,
										doc.WordsFreq,
									))
								}
							}
							e.rworker.Request <- &ranker.RankerRequest {
								QueryId    : sresponse.QueryId,
								Query      : sresponse.Query,
								WordsNum   : sresponse.WordsNum,
								Words      : sresponse.Words,
								WordDocNum : sresponse.WordDocNum,
								RankDocs   : rankDocs,
							}
							rresponse := <- e.rworker.Respone
							if len(rresponse.RetDocs) != 0 {
								//回调返回
								return queryReq.retCall(rresponse.RetDocs)
							}
						}
					} else {
						empty := make([]*include.RetDocument, 0)
						return queryReq.retCall(empty)
					}
				}
			}

			return nil, nil
		},
	}

	return
}