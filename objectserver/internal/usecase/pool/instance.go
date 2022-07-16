package pool

import (
	"common/cache"
	"objectserver/config"
	"time"

	"github.com/838239178/goodmq"
	"github.com/allegro/bigcache"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	Config *config.Config
	Cache  cache.ICache
	Etcd   *clientv3.Client
	Amqp   *goodmq.AmqpConnection
)

func InitPool(cfg *config.Config) {
	Config = cfg

	goodmq.RecoverDelay = 3 * time.Second
	Amqp = goodmq.NewAmqpConnection(cfg.AmqpAddress)
	//init cache
	{
		cacheConf := bigcache.DefaultConfig(cfg.Cache.TTL)
		cacheConf.CleanWindow = cfg.Cache.CleanInterval
		cacheConf.HardMaxCacheSize = int(cfg.Cache.MaxSizeMB)
		cacheConf.Shards = (cfg.Cache.MaxSizeMB / cfg.Cache.MaxItemSizeMB).IntValue()
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
}

func Close() {
	Amqp.Close()
}
