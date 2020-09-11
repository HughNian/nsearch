package engine

import (
	"nsearch/indexer"
	"nsearch/parter"
	"nsearch/constant"
	"nsearch/ranker"
	"nsearch/utils"
	"nsearch/include"
	"sync"
	"log"
	"strings"
	"fmt"
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
			}
		})

		//运行分词器
		go eng.pworker.DoParter()
		//运行搜索器，添加索引
		go eng.iworker.AddIndex()
		//运行索引器，搜索索引
		go eng.iworker.FindIndex()
		//运行排序器
		go eng.rworker.DocRank()
	}

	return eng
}

//把内容加入索引
func (e *Engine) IndexDoc(docId, docType int, content string) {
	if !e.inited {
		log.Fatal("搜索引擎必须初始化")
	}

	if len(content) == 0 {
		log.Fatalln("缺少索引内容")
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
							Update   : false,
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


}

//搜索
func (e *Engine) NSearch(query string, retCall RetCall) {
	if !e.inited {
		log.Fatal("搜索引擎必须初始化")
		return
	}

	if len(query) == 0 {
		log.Fatalln("缺少搜索内容")
		return
	}

	//搜索查询请求
	queryReq := &QueryRequest{
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
						qid      int
						useWords []string
					)
					for _, word := range words {
						if len(word) != 0 && !e.stopWords.StopWordsExist(word) {
							useWords = append(useWords, word)
						}
					}

					qid = utils.GetKeysId(useWords...)
					e.iworker.Srequest <- &indexer.SearchRequest {
						QueryId  : qid,
						Query    : query,
						WordsNum : float32(wordsNum),
						Words    : useWords,
					}

					sresponse := <- e.iworker.Srespone
					if len(sresponse.InterDocs) > 0 {
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