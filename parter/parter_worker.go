package parter

import (
	cli "github.com/HughNian/nmid/pkg/client"
	"github.com/HughNian/nmid/pkg/model"
	"github.com/vmihailenco/msgpack"
	"io"
	"log"
	"nsearch/constant"
	"strings"
	"sync"
)

const MAX_CONTENT_SIZE = 512

// 分词器工作协程
var (
	parter *ParterWorker
	once   sync.Once
)

// 索引分词结果
type PaterResult func(data []byte) (interface{}, error)

// 查询分词结果
type QueryPaterResult func(data []byte) (interface{}, error)

type ParterWorker struct {
	inited bool

	NmidClient  *cli.Client
	nmidSerAddr string

	Request chan *PaterRequest
}

type PaterRequest struct {
	ParterMode string //分词模式
	ParterType int    //分词请求类型: 1-文档分词，2-query查询分词
	ParterTag  string //分词信息展示: "0"-不展示,"1"-显示词的词性,"2"-显示词的词频、distance

	DocId   int
	DocType int
	Content string

	Result  PaterResult
	QResult QueryPaterResult
}

func NewParterWorker() *ParterWorker {
	if parter == nil {
		once.Do(func() {
			serverAddr := strings.Join([]string{constant.NPW_NMID_SERVER_HOST, constant.NPW_NMID_SERVER_PORT}, ":")
			client, err := cli.NewClient("tcp", serverAddr)
			if err == nil {
				parter = &ParterWorker{
					inited:      true,
					NmidClient:  client,
					nmidSerAddr: serverAddr,
					Request:     make(chan *PaterRequest, constant.CHAN_SIZE),
				}
			}
		})
	}

	if parter != nil && parter.inited == true {
		parter.NmidClient.ErrHandler = func(e error) {
			if model.RESTIMEOUT == e {
				parter.inited = false
				parter = nil
			} else if io.EOF == e {
				parter.inited = false
				parter = nil
			}
		}
	}

	return parter
}

func (pr *PaterRequest) RespHandler(resp *cli.Response) {
	if resp.DataType == model.PDT_S_RETURN_DATA && resp.RetLen != 0 {
		if resp.RetLen == 0 {
			log.Println("ret empty")
			return
		}

		var retStruct model.RetStruct
		err := msgpack.Unmarshal(resp.Ret, &retStruct)
		if nil != err {
			log.Fatalln(err)
			return
		}

		if retStruct.Code != 0 {
			log.Println(retStruct.Msg)
			return
		}

		if pr.Result != nil {
			(pr.Result)(retStruct.Data)
		}
	}
}

func (pr *PaterRequest) QRespHandler(resp *cli.Response) {
	if resp.DataType == model.PDT_S_RETURN_DATA && resp.RetLen != 0 {
		if resp.RetLen == 0 {
			log.Println("ret empty")
			return
		}

		var retStruct model.RetStruct
		err := msgpack.Unmarshal(resp.Ret, &retStruct)
		if nil != err {
			log.Fatalln(err)
			return
		}

		if retStruct.Code != 0 {
			log.Println(retStruct.Msg)
			return
		}

		if pr.QResult != nil {
			(pr.QResult)(retStruct.Data)
		}
	}
}

func (pr *PaterRequest) PartWords(mode, text, tag string) error {
	if parter == nil && parter.inited == false {
		parter = NewParterWorker()
	}

	ptext := make(map[string]interface{})
	ptext["text"] = text
	ptext["p2"] = tag
	params, err := msgpack.Marshal(&ptext)
	if err == nil {
		if pr.ParterType == constant.PARTER_TYPE_ONE {
			return parter.NmidClient.Do(mode, params, pr.RespHandler)
		} else if pr.ParterType == constant.PARTER_TYPE_TWO {
			return parter.NmidClient.Do(mode, params, pr.QRespHandler)
		}
	}

	return err
}

func (pw *ParterWorker) DoParter() {
	for true {
		request := <-pw.Request

		if len(request.Content) != 0 {
			//对文档分词
			if request.ParterType == constant.PARTER_TYPE_ONE {
				if string(request.ParterTag) != "1" || string(request.ParterTag) != "2" {
					request.ParterTag = "2"
				}

				//如果文档内容过大则分段分词
				crune := []rune(request.Content)
				clen := len(crune)
				if clen > MAX_CONTENT_SIZE {
					for n := 0; n < clen; n += MAX_CONTENT_SIZE {
						start := n
						end := start + MAX_CONTENT_SIZE
						if end > clen {
							end = clen
						}

						request.PartWords(request.ParterMode, string(crune[start:end]), string(request.ParterTag))
					}
				} else {
					request.PartWords(request.ParterMode, request.Content, string(request.ParterTag))
				}
			}

			//对查询分词
			if request.ParterType == constant.PARTER_TYPE_TWO {
				if string(request.ParterTag) != "0" {
					request.ParterTag = "0"
				}

				if len(request.Content) > 0 {
					request.PartWords(request.ParterMode, request.Content, string(request.ParterTag))
				}
			}
		}
	}
}
