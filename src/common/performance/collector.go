package performance

import (
	"sync"
	"time"
)

var (
	localStore  Store
	remoteStore Store
)

func SetLocalStore(s Store) {
	localStore = s
}

func SetRemoteStore(s Store) {
	remoteStore = s
}

func getStore(st StoreType) Store {
	switch st {
	case None:
		return NoneStore()
	case Local:
		if localStore == nil {
			panic("local store not set")
		}
		return localStore
	case Remote:
		if remoteStore == nil {
			panic("remote store not set")
		}
		return remoteStore
	default:
		panic("invalid store type")
	}
}

type Collector struct {
	conf         *Config
	store        Store
	memData      []*Perform
	lastSaveTime time.Time
	mux          sync.Locker
}

func NewCollector(cfg *Config) *Collector {
	return &Collector{
		conf:         cfg,
		store:        getStore(cfg.Store),
		memData:      make([]*Perform, 0, cfg.MaxInMemeory),
		lastSaveTime: time.Time{},
		mux:          &sync.Mutex{},
	}
}
