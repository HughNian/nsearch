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
	text := []string{"1", "2", "明朝翰林院以永乐朝为界"}
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
	text2 := []string{"2", "2", "明朝翰林院内阁大学士张居正"}
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
	text3 := []string{"3", "2", "明朝永乐时期内阁首辅谢晋"}
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
	text4 := []string{"4", "2", "距今300年的明朝翰林院开启了我国内阁制度，取消了宰相制度"}
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
	text5 := []string{"5", "2", "大明正德翰林院大学士杨廷和"}
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
	text6 := []string{"6", "2", "大学士杨廷和是明朝大才子杨慎的父亲"}
	params6, err := msgpack.Marshal(&text6)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params6, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	text7 := []string{"7", "2", "翰林院文渊阁大学士杨廷和"}
	params7, err := msgpack.Marshal(&text7)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("IndexDoc", params7, respHandlerIndex)
	if nil != err {
		fmt.Println(err)
	}

	text8 := []string{"8", "2", "翰林院大学士杨廷和为大明内阁首辅"}
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
	stext := []string{`大学士`}
	params9, err := msgpack.Marshal(&stext)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("NSearch", params9, respHandlerSearch)
	if nil != err {
		fmt.Println(err)
	}

	stext2 := []string{`明朝张居正`}
	params10, err := msgpack.Marshal(&stext2)
	if err != nil {
		log.Fatalln("params msgpack error:", err)
		os.Exit(1)
	}
	err = client.Do("NSearch", params10, respHandlerSearch)
	if nil != err {
		fmt.Println(err)
	}
}