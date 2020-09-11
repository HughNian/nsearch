## nsearch全文搜索引擎系统

nsearch：golang实现全文搜索引擎系统。基于npartword分词系统进行分词。分词系统与搜索系统的通信都是基于nmid的worker端和client端。
nsearch作为搜索服务的话，也是基于nmid作为worker端。本系统采用了btree的倒排索引，文本相关度算法的，tf-idf，bm25等算法。

启动搜索引擎：   
1.先确保您的nmid调度系统已运行。  

2.确保您的npartword分词系统已运行。  
    
3.make编译完成运行可以执行文件  

使用搜索引擎：  
1.首先你需要往搜索引擎中加入索引内容，索引引擎需要内容，才能有搜索结果。  
   
2.搜索你需要的结果，具体的使用参考client目录中的代码。      
