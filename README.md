<div align="center">
    <a href="http://www.niansong.top"><img src="https://raw.githubusercontent.com/HughNian/nsearch/master/nsearch_logo.png" alt="nsearch logo" width="360"></a>
</div>

## nsearch全文搜索引擎系统

nsearch：golang实现全文搜索引擎系统。基于npartword分词系统进行分词。分词系统与搜索系统的通信都是基于nmid的worker端和client端。   

nsearch作为搜索服务的话，也是基于nmid作为worker端。本系统采用了btree的倒排索引，文本相关度算法的，tf-idf，bm25等算法。    

加入了bolt、sqlite3作为持久层存储，防止由于索引数据过大一直占用内存空间。   


## 使用
启动搜索引擎：   
1.先确保您的nmid调度系统已运行。  

2.确保您的npartword分词系统已运行。  
    
3.make编译完成运行可以执行文件  

使用搜索引擎：  
1.首先你需要往搜索引擎中加入索引内容，索引引擎需要内容，才能有搜索结果。  
   
2.搜索你需要的结果，具体的使用参考client目录中的代码。  

## 示例
```php
//golang调用示例
package main

import (
	"github.com/vmihailenco/msgpack"
	cli "github.com/HughNian/nmid/client"
	"log"
	"fmt"
	"os"
)

const SERVERHOST = "xxx.xxx.x.xxx"
const SERVERPORT = "xxxx"

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
    
    //添加、更新索引
    docId   := "1"
    docType := "1"
    content := "文本"
    text := []string{docId, docType, content}
    params, err := msgpack.Marshal(&text)
    if err != nil {
        log.Fatalln("params msgpack error:", err)
        os.Exit(1)
    }
    err = client.Do("IndexDoc", params, respHandlerIndex)
    if nil != err {
        fmt.Println(err)
    }
    
    //删除索引
    docId   := "1"
    docType := "1"
    content := "文本"
    text := []string{docId, docType, content}
    params, err := msgpack.Marshal(&text)
    if err != nil {
        log.Fatalln("params msgpack error:", err)
        os.Exit(1)
    }
    err = client.Do("DelIndexDoc", params, respHandlerIndex)
    if nil != err {
        fmt.Println(err)
    }
    
    //刷新索引
    text := []string{"1"}
    params, err := msgpack.Marshal(&text)
    if err != nil {
        log.Fatalln("params msgpack error:", err)
        os.Exit(1)
    }
    err = client.Do("FlushIndex", params, respHandlerIndex)
    if nil != err {
        fmt.Println(err)
    }       
    
    //全文搜索
    query := "xxx"; //关键词
    mode  := "1";   //搜索精准度 1-模糊，2-精准
    page  := "1";   //结果分页
    limit := "10";  //分页展示条数
    stext := []string{query, mode, page, limit}
    params9, err := msgpack.Marshal(&stext)
    if err != nil {
        log.Fatalln("params msgpack error:", err)
        os.Exit(1)
    }
    err = client.Do("NSearch", params9, respHandlerSearch)
    if nil != err {
        fmt.Println(err)
    }
}
```

```php
php调用示例

$host = 'xxx.xxx.x.xx';
$port = xx;

$this->client = new ClientExt($host, $port);
$this->client->connect();

$docId = 1;       //自定义内容id
$docType = 1;     //自定义内容类型type值
$content = "文本" //自定义内容

//添加、更新索引
$params = msgpack_pack(array("{$docId}", "{$docType}", "{$content}"));
$return = array();
$this->client->dowork("IndexDoc", $params, function($ret) use (&$return) {
    if($ret[0] != 0) {
        $return = "error";
    } else {
        $return = $ret[1];
    }
});

//删除索引
$params = msgpack_pack(array("{$docId}", "{$docType}", "{$content}"));
$return = array();
$this->client->dowork("DelIndexDoc", $params, function($ret) use (&$return) {
    if($ret[0] != 0) {
        $return ="error";
    } else {
        $return = $ret[1];
    }
});

//刷新索引
$params = msgpack_pack(array("1"));
$return = array();
$this->client->dowork("FlushIndex", $params, function($ret) use (&$return) {
    if($ret[0] != 0) {
        $return = "error";
    } else {
        $return = $ret[1];
    }
});

//全文搜索
$query = "xxx"; //关键词
$mode  = "1";   //搜索精准度 1-模糊，2-精准
$page  = 1;     //结果分页
$limit = 10;    //分页展示条数
$params = msgpack_pack(array("{$query}", $mode, "{$page}", "{$limit}"));
$return = array();
$this->client->dowork("NSearch", $params, function($ret) use (&$return) {
    if($ret[0] != 0) {
        $return = "error";
    } else {
        $return = $ret[2];
    }
});
```

todo   
增加词向量，联想功能。如“普洗”可以联想到“洗车”，“精洗”等。  

## 交流博客
http://www.niansong.top
