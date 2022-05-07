package main

import (
	"fmt"
	"github.com/838239178/goodmq"
	"github.com/allegro/bigcache"
	"github.com/gin-gonic/gin"
	"goodfs/lib/util/cache"
	"goodfs/objectserver/config"
	"goodfs/objectserver/controller"
	"goodfs/objectserver/controller/heartbeat"
	"goodfs/objectserver/controller/locate"
	"goodfs/objectserver/controller/temp"
	"goodfs/objectserver/global"
	"goodfs/objectserver/service"
	"os"
)

func initialize() {
	global.Config = config.ReadConfig()
	//init amqp
	{
		hn, e := os.Hostname()
		if e != nil {
			panic(e)
		}
		config.LocalAddr = fmt.Sprintf("%v:%v", hn, global.Config.Port)
		global.AmqpConnection = goodmq.NewAmqpConnection(global.Config.AmqpAddress)
	}
	//init cache
	{
		cacheConf := bigcache.DefaultConfig(global.Config.Cache.TTL)
		cacheConf.CleanWindow = global.Config.Cache.CleanInterval
		cacheConf.HardMaxCacheSize = int(global.Config.Cache.MaxSizeMB)
		cacheConf.Shards = (global.Config.Cache.MaxSizeMB / global.Config.Cache.MaxItemSizeMB).IntValue()
		global.Cache = cache.NewCache(cacheConf)
	}
	{
		if !service.ExistPath(global.Config.TempPath) {
			if e := os.Mkdir(global.Config.TempPath, os.ModeDir); e != nil {
				panic(e)
			}
		}
		if !service.ExistPath(global.Config.StoragePath) {
			if e := os.Mkdir(global.Config.StoragePath, os.ModeDir); e != nil {
				panic(e)
			}
		}
	}
}

func shutdown() {
	global.Cache.Close()
	err := global.AmqpConnection.Close()
	if err != nil {
		panic(err)
	}
}

func main() {
	initialize()
	defer shutdown()

	locate.WarmUpLocateCache()

	go temp.HandleTempRemove(global.Cache.NotifyEvicted())
	go heartbeat.StartHeartbeat()
	go locate.StartLocate()

	router := gin.Default()

	//init router
	controller.Router(router)

	err := router.Run(":" + global.Config.Port)
	if err != nil {
		panic(err)
	}
}
