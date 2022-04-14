package main

import (
	"fmt"
	"goodfs/objectserver/config"
	"goodfs/objectserver/controller"
	"goodfs/objectserver/controller/heartbeat"
	"goodfs/objectserver/controller/locate"
	"goodfs/objectserver/global"
	"os"
	"strconv"

	"github.com/838239178/goodmq"
	"github.com/VictoriaMetrics/fastcache"
	"github.com/gin-gonic/gin"
)

func initialize() {
	hn, e := os.Hostname()
	if e != nil {
		panic(e)
	}
	config.LocalAddr = fmt.Sprintf("%v:%v", hn, config.Port)
	global.AmqpConnection = goodmq.NewAmqpConnection(config.AmqpAddress)
	global.Cache = fastcache.New(config.CacheSize.IntValue())
}

func close() {
	global.AmqpConnection.Close()
	global.Cache.Reset()
}

func main() {
	initialize()
	defer close()

	go heartbeat.StartHeartbeat()
	go locate.StartLocate()

	router := gin.Default()

	//init router
	controller.Router(router)

	router.Run(":" + strconv.Itoa(config.Port))
}
