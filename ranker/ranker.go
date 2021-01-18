package ranker

import (
	"nsearch/constant"
	"nsearch/include"
	"sort"
	"math"
	"fmt"
)

const (
	K1 = 2.0
	B  = 0.75
)

type queryRet func(query *RankerQuery)

//排序器
type Rank struct {
	queryNum        int
	querys          chan *RankerQuery
	result          queryRet

	DocAllNum       uint64  //索引文档总数
	DocAllWordsNum  float32 //索引文档总的关键词数
}

//排序查询
type RankerQuery struct {
	done          bool

	queryId       uint64                   //查询id
	query         string                   //查询短语
	wordsNum      float32                  //查询分词数
	words         []string                 //查询分词
	wordDocNum    map[string]float32       //查询分词出现该词的文档数
	rankDocs      rankDoc                  //排序文档
	retDocs       []*include.RetDocument   //结果文档
}

//搜索排序文档
type RankDocument struct {
	DocId       int
	DocType     int
	ContentLen  int
	Content     string

	WordsNum    float32             //该文档总词数
	Words       []string            //该文档所有分词
	WordsLen    int                 //该文档词的总长度
	WordsTime   map[string]int      //该文档所有分词在该文档中出现的次数
	WordsFreq   map[string]float64  //该文档所有分词在该文档中的词频

	Bm25        float32             //文档bm25值
}

type rankDoc []*RankDocument

func (rd rankDoc) Len() int {
	return len(rd)
}

func (rd rankDoc) Less(i, j int) bool {
	if rd[i].Bm25 == rd[j].Bm25 {
		return rd[i].WordsNum < rd[j].WordsNum
	}

	return rd[i].Bm25 < rd[j].Bm25
}

func (rd rankDoc) Swap(i, j int) {
	rd[i], rd[j] = rd[j], rd[i]
}

func NewRank() *Rank {
	rank := &Rank {
		queryNum  : 0,
		querys    : make(chan *RankerQuery, constant.CHAN_SIZE),
		DocAllNum : uint64(0),
		DocAllWordsNum : float32(0),
	}

	return rank
}

func (r *Rank) DoRank() {
	for {
		querys := <- r.querys

		if !querys.done {
			avgDocLen := r.DocAllWordsNum / float32(r.DocAllNum) //平均文本关键词长度，用于计算BM25

			for _, doc := range querys.rankDocs {
				doc.Bm25 = r.getBm25Val(querys, doc, avgDocLen)
			}
			sort.Sort(sort.Reverse(querys.rankDocs))

			for _, doc2 := range querys.rankDocs {
				fmt.Println("content:", doc2.Content, "bm25:", doc2.Bm25)
			}

			retDocs := make([]*include.RetDocument, len(querys.rankDocs))
			for k, doc := range querys.rankDocs {
				retDocs[k] = &include.RetDocument {
					DocId   : doc.DocId,
					DocType : doc.DocType,
					ContentLen : doc.ContentLen,
					Content : doc.Content,
				}
			}
			querys.retDocs = retDocs
			r.result(querys)
		}
	}
}

//获取文档bm25值
//@params querys    - 排序查询
//@params doc       - 排序文档
//@params avgDocLen - 平均文本关键词长度
func (r *Rank) getBm25Val(querys *RankerQuery, doc *RankDocument, avgDocLen float32) float32 {
	bm25 := float32(0)
	qwords := querys.words

	if len(qwords) > 0 {
		for _, w := range qwords {
			idf := float32(math.Log2(float64(r.DocAllNum) / float64(querys.wordDocNum[w]) + 1))
			tf  := float32(doc.WordsFreq[w])
			d   := doc.WordsNum

			val := idf * tf * (K1 + 1) / (tf + K1 * (1 - B + B * d / avgDocLen))
			if math.IsNaN(float64(val)) {
				val = float32(0)
			}
			bm25 += val
		}
	}

	return bm25
}