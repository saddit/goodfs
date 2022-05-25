package pool

import (
	"common/cache"
	"objectserver/config"
	"time"

	"github.com/838239178/goodmq"
	"github.com/allegro/bigcache"
)

var (
	Config *config.Config
	Cache  cache.ICache
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
}

func Close() {
	Amqp.Close()
}
