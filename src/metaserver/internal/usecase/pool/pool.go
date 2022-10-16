package pool

import (
	"common/cache"
	"common/constrant"
	"common/etcd"
	"common/registry"
	"common/util"
	"fmt"
	"metaserver/config"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/raftimpl"
	"time"

	"github.com/allegro/bigcache"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	Config       *config.Config
	Cache        cache.ICache
	Storage      *db.Storage
	HashSlot     *db.HashSlotDB
	RaftWrapper  *raftimpl.RaftWrapper
	Etcd         *clientv3.Client
	Registry     *registry.EtcdRegistry
	HttpHostPort string
	GrpcHostPort string
)

func InitPool(cfg *config.Config) {
	Config = cfg
	HttpHostPort = util.GetHostPort(cfg.Port)
	GrpcHostPort = util.GetHostPort(cfg.Cluster.Port)
	initCache(cfg.Cache)
	initEtcd(&cfg.Etcd)
	initRegistry(&cfg.Registry, Etcd, HttpHostPort)
	initStorage(cfg)
	initHashSlot(&cfg.Registry, Etcd)
}

func initEtcd(cfg *etcd.Config) {
	var err error
	Etcd, err = clientv3.New(clientv3.Config{
		Endpoints: cfg.Endpoint,
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
	if err != nil {
		panic(fmt.Errorf("create etcd client err: %v", err))
	}
}

func initStorage(cfg *config.Config) {
	// open db file
	Storage = db.NewStorage()
	if err := Storage.Open(cfg.DataDir); err != nil {
		panic(fmt.Errorf("open db err: %v", err))
	}
}

func initRegistry(cfg *registry.Config, etcd *clientv3.Client, addr string) {
	Registry = registry.NewEtcdRegistry(etcd, *cfg, addr)
}

func initHashSlot(cfg *registry.Config, etcd *clientv3.Client) {
	HashSlot = db.NewHashSlotDB(constrant.EtcdPrefix.FmtHashSlot(cfg.Group, cfg.Name, ""), etcd)
}

func initCache(cfg config.CacheConfig) {
	conf := bigcache.DefaultConfig(cfg.TTL)
	conf.HardMaxCacheSize = int(cfg.MaxSize.MegaByte())
	conf.CleanWindow = cfg.CleanInterval
	Cache = cache.NewCache(conf)
}

func Close() {
	util.LogErr(Storage.Stop())
	util.LogErr(Etcd.Close())
	util.LogErr(HashSlot.Close(time.Minute))
	if RaftWrapper != nil {
		util.LogErr(RaftWrapper.Close())
	}
}
