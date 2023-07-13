package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	cli "github.com/HughNian/nmid/pkg/client"
	"github.com/HughNian/nmid/pkg/model"
	"github.com/vmihailenco/msgpack"
)

var once sync.Once
var client *cli.Client
var err error

const NMIDSERVERHOST = "127.0.0.1"
const NMIDSERVERPORT = "6808"

func getClient() *cli.Client {
	serverAddr := NMIDSERVERHOST + ":" + NMIDSERVERPORT
	client, err = cli.NewClient("tcp", serverAddr).Start()
	if nil == client || err != nil {
		log.Println(err)
	}

	return client
}

func main() {
	var client *cli.Client
	var err error

	client = getClient()
	client.SetParamsType(model.PARAMS_TYPE_JSON)

	client.ErrHandler = func(e error) {
		if model.RESTIMEOUT == e {
			log.Println("time out here")
		} else {
			log.Println(e)
		}
		fmt.Println("client err here")
	}

	respHandlerIndex := func(resp *cli.Response) {
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

			fmt.Println(retStruct.Msg)
			return
		}
	}

	respHandlerSearch := func(resp *cli.Response) {
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
				fmt.Println(retStruct.Msg)
				return
			}

			fmt.Println(string(retStruct.Data))
			fmt.Print("\n\n")
		}
	}

	/***add index***/
	//1
	//text := []string{"1", "3", "【正品授权】壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 0W-30 SN 灰壳（4L装）"}
	//params, err := msgpack.Marshal(&text)

	paramsName1 := make(map[string]interface{})
	paramsName1["docId"] = "1"
	paramsName1["docType"] = "3"
	paramsName1["content"] = "【正品授权】壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 0W-30 SN 灰壳（4L装）"
	params1, err := json.Marshal(&paramsName1)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params1, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	//2
	//text2 := []string{"2", "3", "【品牌直供】嘉实多/Castrol 金嘉护机油 5W-40 SN级 合成技术（4L装）"}
	//params2, err := msgpack.Marshal(&text2)

	paramsName2 := make(map[string]interface{})
	paramsName2["docId"] = "2"
	paramsName2["docType"] = "3"
	paramsName2["content"] = "【品牌直供】嘉实多/Castrol 金嘉护机油 5W-40 SN级 合成技术（4L装）"
	params2, err := json.Marshal(&paramsName2)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params2, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	//3
	//text3 := []string{"3", "3", "【正品授权】壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 5W-40 SN 灰壳（4L装）"}
	//params3, err := msgpack.Marshal(&text3)

	paramsName3 := make(map[string]interface{})
	paramsName3["docId"] = "3"
	paramsName3["docType"] = "3"
	paramsName3["content"] = "【正品授权】壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 5W-40 SN 灰壳（4L装）"
	params3, err := json.Marshal(&paramsName3)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params3, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	//4
	//text4 := []string{"4", "3", "【正品授权】壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 5W-40 SN 灰壳（4L装）"}
	//params4, err := msgpack.Marshal(&text4)

	paramsName4 := make(map[string]interface{})
	paramsName4["docId"] = "4"
	paramsName4["docType"] = "3"
	paramsName4["content"] = "【正品授权】壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 5W-40 SN 灰壳（4L装）"
	params4, err := json.Marshal(&paramsName4)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params4, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	//5
	//text5 := []string{"5", "3", "【正品授权】美孚/Mobil 新速霸1000合成机油 5W-30 SN级 4L"}
	//params5, err := msgpack.Marshal(&text5)

	paramsName5 := make(map[string]interface{})
	paramsName5["docId"] = "5"
	paramsName5["docType"] = "3"
	paramsName5["content"] = "【正品授权】美孚/Mobil 新速霸1000合成机油 5W-30 SN级 4L"
	params5, err := json.Marshal(&paramsName5)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params5, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	//6
	//text6 := []string{"6", "3", "【品牌直供】嘉实多/Castrol 金嘉护机油 10W-40 SN级 合成技术（4L装）"}
	//params6, err := msgpack.Marshal(&text6)

	paramsName6 := make(map[string]interface{})
	paramsName6["docId"] = "6"
	paramsName6["docType"] = "3"
	paramsName6["content"] = "【品牌直供】嘉实多/Castrol 金嘉护机油 10W-40 SN级 合成技术（4L装）"
	params6, err := json.Marshal(&paramsName6)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params6, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	//text7 := []string{"7", "3", "【正品授权】壳牌/Shell 喜力半合成机油HX7 5W-40 SN/CF 蓝壳（4L装）"}
	//params7, err := msgpack.Marshal(&text7)

	paramsName7 := make(map[string]interface{})
	paramsName7["docId"] = "7"
	paramsName7["docType"] = "3"
	paramsName7["content"] = "【正品授权】壳牌/Shell 喜力半合成机油HX7 5W-40 SN/CF 蓝壳（4L装）"
	params7, err := json.Marshal(&paramsName7)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params7, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	//text8 := []string{"8", "3", "【正品授权】壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 5W-30 SN 灰壳（4L装）"}
	//params8, err := msgpack.Marshal(&text8)

	paramsName8 := make(map[string]interface{})
	paramsName8["docId"] = "8"
	paramsName8["docType"] = "3"
	paramsName8["content"] = "【正品授权】壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 5W-30 SN 灰壳（4L装）"
	params8, err := json.Marshal(&paramsName8)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params8, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	/***search***/
	//stext := []string{`壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 5W-30 SN 灰壳（4L装）`, "1", "1", "10"}
	//params9, err := msgpack.Marshal(&stext)

	paramsName9 := make(map[string]interface{})
	paramsName9["query"] = "壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 5W-30 SN 灰壳（4L装）"
	paramsName9["mode"] = "1"
	paramsName9["page"] = "1"
	paramsName9["limit"] = "10"
	params9, err := json.Marshal(&paramsName9)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("NSearch", params9, respHandlerSearch)
	if nil != err {
		fmt.Println(err)
	}

	//stext2 := []string{`10W-40 SN`, "1", "1", "10"}
	//params10, err := msgpack.Marshal(&stext2)

	paramsName10 := make(map[string]interface{})
	paramsName10["query"] = "10W-40 SN"
	paramsName10["mode"] = "1"
	paramsName10["page"] = "1"
	paramsName10["limit"] = "10"
	params10, err := json.Marshal(&paramsName10)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("NSearch", params10, respHandlerSearch)
	if nil != err {
		fmt.Println(err)
	}

	//stext3 := []string{`5W-30 SN`, "1", "1", "10"}
	//params11, err := msgpack.Marshal(&stext3)

	paramsName11 := make(map[string]interface{})
	paramsName11["query"] = "5W-30 SN"
	paramsName11["mode"] = "1"
	paramsName11["page"] = "1"
	paramsName11["limit"] = "10"
	params11, err := json.Marshal(&paramsName11)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("NSearch", params11, respHandlerSearch)
	if nil != err {
		fmt.Println(err)
	}

	//stext4 := []string{`灰壳 5W-40`, "1", "1", "10"}
	//params12, err := msgpack.Marshal(&stext4)

	paramsName12 := make(map[string]interface{})
	paramsName12["query"] = "灰壳 5W-40"
	paramsName12["mode"] = "1"
	paramsName12["page"] = "1"
	paramsName12["limit"] = "10"
	params12, err := json.Marshal(&paramsName12)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("NSearch", params12, respHandlerSearch)
	if nil != err {
		fmt.Println(err)
	}
}
