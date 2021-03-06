package constant

const CHAN_SIZE = 64

//文档类型
const (
	NORMAL   = 1 //文章
	GOODS    = 2 //商品
	SERVICES = 3 //服务
	STAND    = 4 //标品
	SHOP     = 5 //店铺
)

//nmid调度服务
const (
	NMID_SERVER_HOST = "192.168.1.176"
	NMID_SERVER_PORT = "6808"
)

//分词类型方法
const (
	PART_MODE_ONE   = "PartWordsM1"  //普通分词方法
	PART_MODE_TWO   = "PartWordsM2"  //mmseg分词方法
	PART_MODE_THREE = "PartWordsM3"  //隐马尔可夫模型分词方法
)

const (
	PARTER_TYPE_ONE = 1 //分词请求类型：1-文档分词
	PARTER_TYPE_TWO = 2 //分词请求类型：2-query查询分词
)