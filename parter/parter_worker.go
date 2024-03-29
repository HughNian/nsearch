package parter

import (
	"log"
	"nsearch/constant"
	"strings"
	"sync"

	cli "github.com/HughNian/nmid/pkg/client"
	"github.com/HughNian/nmid/pkg/logger"
	"github.com/HughNian/nmid/pkg/model"
	wor "github.com/HughNian/nmid/pkg/worker"
	"github.com/vmihailenco/msgpack"
)

const MAX_CONTENT_SIZE = 512

// 分词器工作协程
var (
	Parter *ParterWorker

	once   sync.Once
	Client *cli.Client
	err    error
)

// PaterResult 索引分词结果
type PaterResult func(data []byte) (interface{}, error)

// QueryPaterResult 查询分词结果
type QueryPaterResult func(data []byte) (interface{}, error)

type ParterWorker struct {
	Request chan *PaterRequest
}

type PaterRequest struct {
	Job wor.Job

	ParterMode string //分词模式
	ParterType int    //分词请求类型: 1-文档分词，2-query查询分词
	ParterTag  string //分词信息展示: "0"-不展示,"1"-显示词的词性,"2"-显示词的词频、distance

	DocId   int
	DocType int
	Content string

	Result  PaterResult
	QResult QueryPaterResult
}

func getClient() *cli.Client {
	serverAddr := strings.Join([]string{constant.NPW_NMID_SERVER_HOST, constant.NPW_NMID_SERVER_PORT}, ":")
	Client, err = cli.NewClient("tcp", serverAddr).Start()
	if nil == Client || err != nil {
		log.Println(err)
	}

	return Client
}

func NewParterWorker() *ParterWorker {
	Parter = &ParterWorker{
		Request: make(chan *PaterRequest, constant.CHAN_SIZE),
	}

	return Parter
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

func (pr *PaterRequest) PartWords(job wor.Job, mode, text, tag string) error {
	client := getClient()
	defer client.Close()

	// errHandler := func(e error) {
	// 	if pr.QResult != nil {
	// 		(pr.QResult)([]byte{})
	// 	}

	// 	if model.RESTIMEOUT == e {
	// 		logger.Info("time out here")
	// 	} else {
	// 		logger.Error(e)
	// 	}
	// }

	client.ErrHandler = func(e error) {
		if pr.QResult != nil {
			(pr.QResult)([]byte{})
		}

		if model.RESTIMEOUT == e {
			logger.Info("time out here")
		} else {
			logger.Error(e)
		}
	}

	//serverAddr := strings.Join([]string{constant.NPW_NMID_SERVER_HOST, constant.NPW_NMID_SERVER_PORT}, ":")

	ptext := make(map[string]interface{})
	ptext["text"] = text
	ptext["p2"] = tag
	params, err := msgpack.Marshal(&ptext)
	if err == nil {
		if pr.ParterType == constant.PARTER_TYPE_ONE {
			return client.Do(mode, params, pr.RespHandler)
			// job.ClientCall(serverAddr, mode, ptext, pr.RespHandler, errHandler)
		} else if pr.ParterType == constant.PARTER_TYPE_TWO {
			return client.Do(mode, params, pr.QRespHandler)
			// job.ClientCall(serverAddr, mode, ptext, pr.QRespHandler, errHandler)
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

						request.PartWords(request.Job, request.ParterMode, string(crune[start:end]), string(request.ParterTag))
					}
				} else {
					request.PartWords(request.Job, request.ParterMode, request.Content, string(request.ParterTag))
				}
			}

			//对查询分词
			if request.ParterType == constant.PARTER_TYPE_TWO {
				if string(request.ParterTag) != "0" {
					request.ParterTag = "0"
				}

				if len(request.Content) > 0 {
					request.PartWords(request.Job, request.ParterMode, request.Content, string(request.ParterTag))
				}
			}
		}
	}
}
