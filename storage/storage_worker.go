package storage

import (
	"sync"
	"nsearch/constant"
)

const (
	BOLT_FILE   = `./nsearch.bolt`
	SQLITE_FILE = `./nsearch.db`
)

//存储器工作协程
var (
	storage  *StorageWorker
	once     sync.Once
)

var engineMap = map[string]func(dbfile string) (Storage, error) {
	"bolt"    : NewBolt,
	"sqlite3" : NewSqlite,
}

type Storage interface {
	AddData(k, v []byte) error
	GetData(k []byte) (v []byte, err error)
	DelData(k []byte) error
	Close() error
}

type StorageWorker struct {
	inited      bool
	engine      string
	Istorage    Storage
	Record      chan map[string][]byte
}

func NewStorageWorker(engine string) *StorageWorker {
	if storage == nil || storage.inited != true {
		once.Do(func() {
			var istorage Storage
			var err error
			if function, has := engineMap[engine]; has {
				if engine == "bolt" {
					istorage, err = function(BOLT_FILE)
				} else if engine == "sqlite3" {
					istorage, err = function(SQLITE_FILE)
				}
				if err != nil {
					return
				}
			} else {
				istorage, err = function(BOLT_FILE)
				if err != nil {
					return
				}
			}

			storage = &StorageWorker {
				inited   : true,
				engine   : engine,
				Istorage : istorage,
				Record   : make(chan map[string][]byte, constant.CHAN_SIZE),
			}
		})
	}

	return storage
}

func (sw *StorageWorker) DoStorage() {
	for true {
		justdoit := <- sw.Record

		if len(justdoit) > 0 {
			for k, v := range justdoit {
				record, err := sw.Istorage.GetData([]byte(k))
				if err == nil {
					nv := append(v, record...) //merge
					sw.Istorage.AddData([]byte(k), nv)
				} else {
					sw.Istorage.AddData([]byte(k), v)
				}
			}
		}
	}
}