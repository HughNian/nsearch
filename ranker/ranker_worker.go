package ranker

import (
	"nsearch/constant"
	"sync"
	"log"
	"nsearch/include"
)

//排序器协程
var (
	ranker *RankerWorker
	once   sync.Once
)

type RankerWorker struct {
	inited  bool

	rank    *Rank
	Request  chan *RankerRequest
	Respone  chan *RankerRespone
}

//查询排序请求
type RankerRequest struct {
	QueryId       uint64              //查询id
	Query         string              //查询短语
	WordsNum      float32             //查询分词数量
	Words         []string            //查询分词
	WordDocNum    map[string]float32  //查询分词出现该词的文档数
	RankDocs      []*RankDocument     //查询排序文档
}

//查询排序结果返回
type RankerRespone struct {
	QueryId     uint64            //查询id
	Query       string            //查询短语
	RetDocs     []*include.RetDocument    //查询结果文档
}

func NewRankerWorker() *RankerWorker {
	if ranker != nil && ranker.inited == true {
		log.Fatal("单实例不能再次初始化")
	}

	if ranker == nil {
		once.Do(func() {
			ranker = &RankerWorker {
				inited   : true,
				rank     : NewRank(),
				Request  : make(chan *RankerRequest, constant.CHAN_SIZE),
				Respone  : make(chan *RankerRespone, constant.CHAN_SIZE),
			}

			go func() {
				ranker.rank.DoRank()
			}()
		})
	}

	return ranker
}

func (rw *RankerWorker) Rank() *Rank {
	return rw.rank
}

func (rw *RankerWorker) NewRankDocument(
	DocId int,
	DocType int,
	ContentLen int,
	Content string,
	WordsNum    float32,
	WordsLen    int,
	Words       []string,
	WordsTime   map[string]int,
	WordsFreq   map[string]float64,
) *RankDocument {
	return &RankDocument {
		DocId      : DocId,
		DocType    : DocType,
		ContentLen : ContentLen,
		Content    : Content,
		WordsNum   : WordsNum,
		WordsLen   : WordsLen,
		Words      : Words,
		WordsTime  : WordsTime,
		WordsFreq  : WordsFreq,
		Bm25       : float32(0),
	}
}

func (rw *RankerWorker) DocRank() {
	for true {
		request := <- rw.Request

		if request.QueryId > 0 && len(request.Query) > 0 {
			rquery := &RankerQuery {
				done      : false,
				queryId   : request.QueryId,
				query     : request.Query,
				wordsNum  : request.WordsNum,
				words     : request.Words,
				wordDocNum : request.WordDocNum,
				rankDocs   : request.RankDocs,
			}
			rw.rank.queryNum++

			rw.rank.querys <- rquery

			rw.rank.result = func(query *RankerQuery) {
				if request.QueryId == query.queryId {
					rw.Respone <- &RankerRespone {
						QueryId : query.queryId,
						Query   : query.query,
						RetDocs : query.retDocs,
					}
				}
			}

		}
	}
}