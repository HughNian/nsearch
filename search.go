package main

import (
	wor "github.com/HughNian/nmid/worker"
	"nsearch/constant"
	"nsearch/engine"
	"log"
	"fmt"
	"github.com/vmihailenco/msgpack"
	"strconv"
	"nsearch/include"
	"encoding/json"
)

var sengine *engine.Engine

func IndexDocment(job wor.Job) ([]byte, error) {
	resp := job.GetResponse()
	if nil == resp {
		return []byte(``), fmt.Errorf("response data error")
	}

	p1   := resp.StrParams[0]
	docId, _  := strconv.ParseInt(p1, 10, 64)
	p2 := resp.StrParams[1]
	docType, _  := strconv.ParseInt(p2, 10, 64)
	content := resp.StrParams[2]
	sengine.IndexDoc(int(docId), int(docType), content)

	retStruct := wor.GetRetStruct()
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

func NSearch(job wor.Job) ([]byte, error) {
	resp := job.GetResponse()
	if nil == resp {
		return []byte(``), fmt.Errorf("response data error")
	}

	resultData := make(chan []byte)
	query := resp.StrParams[0]
	retStruct := wor.GetRetStruct()
	sengine.NSearch(query, func(result []*include.RetDocument) (ret []byte, err error) {
		if result != nil && len(result) != 0 {
			data, err := json.Marshal(result)
			if err == nil {
				fmt.Println("search data:", string(data))
			} else {
				log.Println(err)
			}

			retStruct.Msg = "ok"
			retStruct.Data = data
		} else {
			retStruct.Code = 100
			retStruct.Msg  = "no result"
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
	returnData := <- resultData

	return returnData, nil
}

func main() {
	var worker *wor.Worker
	var err error
	serverAddr := constant.NMID_SERVER_HOST + ":" + constant.NMID_SERVER_PORT
	worker = wor.NewWorker()
	err = worker.AddServer("tcp", serverAddr)
	if err != nil {
		log.Fatalln(err)
		worker.WorkerClose()
		return
	}

	sengine = engine.NewEngine()
	if sengine != nil {
		worker.AddFunction("IndexDoc", IndexDocment)
		worker.AddFunction("NSearch", NSearch)
	}

	if err = worker.WorkerReady(); err != nil {
		log.Fatalln(err)
		worker.WorkerClose()
		return
	}

	worker.WorkerDo()
}