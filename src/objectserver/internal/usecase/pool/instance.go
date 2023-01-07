package pool

import (
	"common/cache"
	"common/etcd"
	"common/registry"
	"common/util"
	"objectserver/config"
	"objectserver/internal/db"
	"objectserver/internal/usecase/service"

	"github.com/allegro/bigcache"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	Config        *config.Config
	Cache         cache.ICache
	Etcd          *clientv3.Client
	Registry      registry.Registry
	Discovery     registry.Discovery
	ObjectCap     *db.ObjectCapacity
	DriverManager *service.DriverManager
)

func InitPool(cfg *config.Config) {
	Config = cfg
	initDm()
	initCache(&cfg.Cache)
	initEtcd(&cfg.Etcd)
	initRegister(Etcd, cfg)
	initObjectCap(Etcd, cfg)
}

func initDm() {
	DriverManager = service.NewDriverManager(service.NewFreeFirstDriver())
}

func initObjectCap(et *clientv3.Client, cfg *config.Config) {
	ObjectCap = db.NewObjectCapacity(et, cfg)
}

func initCache(cfg *config.CacheConfig) {
	cacheConf := bigcache.DefaultConfig(cfg.TTL)
	cacheConf.CleanWindow = cfg.CleanInterval
	cacheConf.HardMaxCacheSize = int(cfg.MaxSize.MegaByte())
	cacheConf.Shards = int(cfg.MaxSize / cfg.MaxItemSize)
	Cache = cache.NewCache(cacheConf)
}

func initEtcd(cfg *etcd.Config) {
	var e error
	if Etcd, e = clientv3.New(clientv3.Config{
		Endpoints: cfg.Endpoint,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}); e != nil {
		panic(e)
	}
}

func initRegister(et *clientv3.Client, cfg *config.Config) {
	cfg.Registry.HttpAddr = util.GetHostPort(cfg.Port)
	cfg.Registry.RpcAddr = util.GetHostPort(cfg.RpcPort)
	er := registry.NewEtcdRegistry(et, cfg.Registry)
	Registry, Discovery = er, er
}

func Close() {
	util.LogErr(Etcd.Close())
	util.LogErr(Cache.Close())
}
