package main

import (
	"github.com/vmihailenco/msgpack"
	cli "github.com/HughNian/nmid/client"
	"log"
	"fmt"
	"os"
)

const SERVERHOST = "192.168.1.176"
const SERVERPORT = "6808"

func main() {
	var client *cli.Client
	var err error

	serverAddr := SERVERHOST + ":" + SERVERPORT
	client, err = cli.NewClient("tcp", serverAddr)
	if nil == client || err != nil {
		log.Println(err)
		return
	}
	defer client.Close()

	client.ErrHandler= func(e error) {
		if cli.RESTIMEOUT == e {
			log.Println("time out here")
		} else {
			log.Println(e)
		}
		fmt.Println("client err here")
	}

	respHandlerIndex := func(resp *cli.Response) {
		if resp.DataType == cli.PDT_S_RETURN_DATA && resp.RetLen != 0 {
			if resp.RetLen == 0 {
				log.Println("ret empty")
				return
			}

			var retStruct cli.RetStruct
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
		if resp.DataType == cli.PDT_S_RETURN_DATA && resp.RetLen != 0 {
			if resp.RetLen == 0 {
				log.Println("ret empty")
				return
			}

			var retStruct cli.RetStruct
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
	text := []string{"1", "3", "【正品授权】壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 0W-30 SN 灰壳（4L装）"}
	params, err := msgpack.Marshal(&text)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	//2
	text2 := []string{"2", "3", "【品牌直供】嘉实多/Castrol 金嘉护机油 5W-40 SN级 合成技术（4L装）"}
	params2, err := msgpack.Marshal(&text2)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params2, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	//3
	text3 := []string{"3", "3", "【正品授权】壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 5W-40 SN 灰壳（4L装）"}
	params3, err := msgpack.Marshal(&text3)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params3, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	//4
	text4 := []string{"4", "3", "【正品授权】壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 5W-40 SN 灰壳（4L装）"}
	params4, err := msgpack.Marshal(&text4)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params4, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	//5
	text5 := []string{"5", "3", "【正品授权】美孚/Mobil 新速霸1000合成机油 5W-30 SN级 4L"}
	params5, err := msgpack.Marshal(&text5)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params5, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	//6
	text6 := []string{"6", "3", "【品牌直供】嘉实多/Castrol 金嘉护机油 10W-40 SN级 合成技术（4L装）"}
	params6, err := msgpack.Marshal(&text6)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params6, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	text7 := []string{"7", "3", "【正品授权】壳牌/Shell 喜力半合成机油HX7 5W-40 SN/CF 蓝壳（4L装）"}
	params7, err := msgpack.Marshal(&text7)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params7, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	text8 := []string{"8", "3", "【正品授权】壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 5W-30 SN 灰壳（4L装）"}
	params8, err := msgpack.Marshal(&text8)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params8, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	/***search***/
	stext := []string{`壳牌/Shell 超凡喜力 全合成机油 新中超版 ULTRA 5W-30 SN 灰壳（4L装）`, "1", "1", "10"}
	params9, err := msgpack.Marshal(&stext)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("NSearch", params9, respHandlerSearch)
	if nil != err {
		fmt.Println(err)
	}

	stext2 := []string{`10W-40 SN`, "1", "1", "10"}
	params10, err := msgpack.Marshal(&stext2)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("NSearch", params10, respHandlerSearch)
	if nil != err {
		fmt.Println(err)
	}

	stext3 := []string{`5W-30 SN`, "1", "1", "10"}
	params11, err := msgpack.Marshal(&stext3)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("NSearch", params11, respHandlerSearch)
	if nil != err {
		fmt.Println(err)
	}

	stext4 := []string{`灰壳 5W-40`, "1", "1", "10"}
	params12, err := msgpack.Marshal(&stext4)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("NSearch", params12, respHandlerSearch)
	if nil != err {
		fmt.Println(err)
	}
}