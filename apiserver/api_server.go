package main

import (
	"github.com/irfansharif/cfilter"
	"goodfs/apiserver/config"
	"goodfs/apiserver/controller"
	"goodfs/apiserver/controller/heartbeat"
	"goodfs/apiserver/controller/locate"
	"goodfs/apiserver/global"
	"log"
	"strconv"
	"time"

	"github.com/838239178/goodmq"
	"github.com/gin-gonic/gin"
)

func initialize() {
	goodmq.RecoverDelay = 3 * time.Second
	global.AmqpConnection = goodmq.NewAmqpConnection(config.AmqpAddress)
	global.ExistFilter = cfilter.New()
}

func shutdown() {
	err := global.AmqpConnection.Close()
	if err != nil {
		log.Println(err)
	}
}

func main() {
	initialize()
	defer shutdown()

	go heartbeat.ListenHeartbeat()
	go locate.SyncExistFilter()

	router := gin.Default()

	api := router.Group("/api")

	controller.Router(api)

	err := router.Run(":" + strconv.Itoa(config.Port))
	if err == nil {
		log.Fatal(err)
	}
}
