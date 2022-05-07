package main

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"goodfs/apiserver/command"
	"goodfs/apiserver/config"
	"goodfs/apiserver/controller"
	"goodfs/apiserver/controller/heartbeat"
	"goodfs/apiserver/global"
	"goodfs/apiserver/service/selector"
	"goodfs/lib/util/datasize"
	"log"
	"net/http"
	"time"

	"github.com/838239178/goodmq"
	"github.com/gin-gonic/gin"
)

func initialize() {
	global.Config = config.ReadConfig()
	global.Http = &http.Client{Timeout: 5 * time.Second}
	goodmq.RecoverDelay = 3 * time.Second
	global.AmqpConnection = goodmq.NewAmqpConnection(global.Config.AmqpAddress)
	global.Balancer = selector.NewSelector(global.Config.SelectStrategy)
	var e error
	if global.LocalDB, e = leveldb.OpenFile(global.Config.LocalStorePath, &opt.Options{
		BlockCacheCapacity:          datasize.MustParse(global.Config.LocalCacheSize).IntValue(),
		CompactionSourceLimitFactor: 5,
	}); e != nil {
		panic(e)
	}

	command.ReadCommand()
}

func shutdown() {
	err := global.AmqpConnection.Close()
	if err != nil {
		log.Println(err)
	}
	global.Http.CloseIdleConnections()
	if err = global.LocalDB.Close(); err != nil {
		log.Println(err)
	}
}

func main() {
	initialize()
	defer shutdown()

	go heartbeat.ListenHeartbeat()

	router := gin.Default()

	api := router.Group("/api")
	controller.Router(api)

	err := router.Run(":" + global.Config.Port)
	if err == nil {
		log.Fatal(err)
	}
}
