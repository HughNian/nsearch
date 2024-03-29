package main

import (
	"encoding/json"
	"fmt"
	"log"
	"nsearch/constant"
	"nsearch/engine"
	"nsearch/include"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/HughNian/nmid/pkg/logger"
	"github.com/HughNian/nmid/pkg/model"
	wor "github.com/HughNian/nmid/pkg/worker"
	"github.com/vmihailenco/msgpack"
)

var sengine *engine.Engine

func IndexDocment(job wor.Job) ([]byte, error) {
	resp := job.GetResponse()
	if nil == resp {
		return []byte(``), fmt.Errorf("response data error")
	}

	docIdStr := resp.ParamsMap["docId"].(string)
	docId, _ := strconv.ParseInt(docIdStr, 10, 64)
	docTypeStr := resp.ParamsMap["docType"].(string)
	docType, _ := strconv.ParseInt(docTypeStr, 10, 64)
	content := resp.ParamsMap["content"].(string)
	sengine.IndexDoc(job, int(docId), int(docType), content)

	retStruct := model.GetRetStruct()
	retStruct.Msg = "add index ok"
	retStruct.Data = []byte(``)
	ret, err := msgpack.Marshal(retStruct)
	if nil != err {
		return []byte(``), err
	}

	resp.RetLen = uint32(len(ret))
	resp.Ret = ret

	return ret, nil
}

func DelIndexDocment(job wor.Job) ([]byte, error) {
	resp := job.GetResponse()
	if nil == resp {
		return []byte(``), fmt.Errorf("response data error")
	}

	docIdStr := resp.ParamsMap["docId"].(string)
	docId, _ := strconv.ParseInt(docIdStr, 10, 64)
	docTypeStr := resp.ParamsMap["docType"].(string)
	docType, _ := strconv.ParseInt(docTypeStr, 10, 64)
	content := resp.ParamsMap["content"].(string)
	sengine.DelIndexDoc(job, int(docId), int(docType), content)

	retStruct := model.GetRetStruct()
	retStruct.Msg = "del index ok"
	retStruct.Data = []byte(``)
	ret, err := msgpack.Marshal(retStruct)
	if nil != err {
		return []byte(``), err
	}

	resp.RetLen = uint32(len(ret))
	resp.Ret = ret

	return ret, nil
}

func NSearch(job wor.Job) ([]byte, error) {
	resp := job.GetResponse()
	if nil == resp {
		return []byte(``), fmt.Errorf("response data error")
	}

	resultData := make(chan []byte)
	query := resp.ParamsMap["query"].(string)
	modeStr := resp.ParamsMap["mode"].(string)
	mode, _ := strconv.ParseInt(modeStr, 10, 64)
	pageStr := resp.ParamsMap["page"].(string)
	page, _ := strconv.ParseInt(pageStr, 10, 64)
	limitStr := resp.ParamsMap["limit"].(string)
	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	retStruct := model.GetRetStruct()
	sengine.NSearch(job, query, int(mode), int(page), int(limit), func(result []*include.RetDocument) (ret []byte, err error) {
		if result != nil && len(result) != 0 {
			data, err := json.Marshal(result)
			if err == nil {
				logger.Infof("search data: %s", string(data))
			} else {
				logger.Error(err)
			}

			retStruct.Msg = "ok"
			retStruct.Data = data
		} else {
			retStruct.Code = 100
			retStruct.Msg = "no result"
			retStruct.Data = []byte(``)
		}

		ret, err = msgpack.Marshal(retStruct)
		if nil != err {
			ret = []byte(``)
			resp.RetLen = uint32(len(ret))
			resp.Ret = ret
			resultData <- ret
			return ret, err
		}

		resp.RetLen = uint32(len(ret))
		resp.Ret = ret
		resultData <- ret
		return ret, nil
	})

	//异步管道通信，在没有数据返回时会阻塞
	returnData := <-resultData

	return returnData, nil
}

func FlushIndex(job wor.Job) ([]byte, error) {
	resp := job.GetResponse()
	if nil == resp {
		return []byte(``), fmt.Errorf("response data error")
	}

	sengine.FlushIndex()

	retStruct := model.GetRetStruct()
	retStruct.Msg = "flush index ok"
	retStruct.Data = []byte(``)
	ret, err := msgpack.Marshal(retStruct)
	if nil != err {
		return []byte(``), err
	}

	resp.RetLen = uint32(len(ret))
	resp.Ret = ret

	return ret, nil
}

func main() {
	showLogo()

	var workerName = "nsearch"
	var worker *wor.Worker
	var err error
	serverAddr := constant.NMID_SERVER_HOST + ":" + constant.NMID_SERVER_PORT
	// worker = wor.NewWorker().SetWorkerName(workerName).WithTrace(constant.SkyWalkingTraceOapUrl)
	worker = wor.NewWorker().SetWorkerName(workerName)
	err = worker.AddServer("tcp", serverAddr)
	if err != nil {
		log.Fatalln(err)
		worker.WorkerClose()
		return
	}

	sengine = engine.NewEngine()
	if sengine != nil {
		worker.AddFunction("IndexDoc", IndexDocment)
		worker.AddFunction("DelIndexDoc", DelIndexDocment)
		worker.AddFunction("NSearch", NSearch)
		worker.AddFunction("FlushIndex", FlushIndex)
	}

	if err = worker.WorkerReady(); err != nil {
		log.Fatalln(err)
		worker.WorkerClose()
		return
	}

	go worker.WorkerDo()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-c
	logger.Infof("Server Exit: %s", s.String())
	worker.WorkerClose()
}

func showLogo() {
	logo := `
                                    __  
   ____  ________  ____ ___________/ /_ 
  / __ \/ ___/ _ \/ __ \/ ___/ ___/ __ \
 / / / (__  )  __/ /_/ / /  / /__/ / / /
/_/ /_/____/\___/\__,_/_/   \___/_/ /_/
`
	fmt.Println(logo)
}
