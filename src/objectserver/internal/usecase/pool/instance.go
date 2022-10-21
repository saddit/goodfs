package pool

import (
	"common/cache"
	"common/registry"
	"common/util"
	"objectserver/config"
	"objectserver/internal/db"

	"github.com/allegro/bigcache"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	Config    *config.Config
	Cache     cache.ICache
	Etcd      *clientv3.Client
	Registry  registry.Registry
	Discovery registry.Discovery
	ObjectCap *db.ObjectCapacity
)

func InitPool(cfg *config.Config) {
	Config = cfg
	//init cache
	{
		cacheConf := bigcache.DefaultConfig(cfg.Cache.TTL)
		cacheConf.CleanWindow = cfg.Cache.CleanInterval
		cacheConf.HardMaxCacheSize = int(cfg.Cache.MaxSize.MegaByte())
		cacheConf.Shards = int(cfg.Cache.MaxSize / cfg.Cache.MaxItemSize)
		Cache = cache.NewCache(cacheConf)
	}

	var e error
	if Etcd, e = clientv3.New(clientv3.Config{
		Endpoints: cfg.Etcd.Endpoint,
		Username:  cfg.Etcd.Username,
		Password:  cfg.Etcd.Password,
	}); e != nil {
		panic(e)
	}

	ObjectCap = db.NewObjectCapacity(Etcd, cfg)
	
	cfg.Registry.HttpAddr = util.GetHostPort(cfg.Port)
	cfg.Registry.RpcAddr = util.GetHostPort(cfg.RpcPort)
	er := registry.NewEtcdRegistry(Etcd, cfg.Registry)
	Registry, Discovery = er, er
}

func Close() {
	util.LogErr(Etcd.Close())
	util.LogErr(Cache.Close())
}
