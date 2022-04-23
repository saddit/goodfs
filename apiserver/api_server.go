package main

import (
	"goodfs/apiserver/command"
	"goodfs/apiserver/config"
	"goodfs/apiserver/controller"
	"goodfs/apiserver/controller/heartbeat"
	"goodfs/apiserver/controller/locate"
	"goodfs/apiserver/global"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/838239178/cfilter"

	"github.com/838239178/goodmq"
	"github.com/gin-gonic/gin"
)

func initialize() {
	global.Http = &http.Client{Timeout: 5 * time.Second}
	goodmq.RecoverDelay = 3 * time.Second
	global.AmqpConnection = goodmq.NewAmqpConnection(config.AmqpAddress)
	global.ExistFilter = cfilter.New()
	
	command.ReadCommand()
}

func shutdown() {
	err := global.AmqpConnection.Close()
	if err != nil {
		log.Println(err)
	}
	global.Http.CloseIdleConnections()
}

func main() {
	initialize()
	defer shutdown()

	go heartbeat.ListenHeartbeat()
	go locate.SyncExistFilter()

	router := gin.Default()

	api := router.Group("/api")
	controller.Router(api)

	help := router.Group("/help")
	controller.HelperRouter(help)

	err := router.Run(":" + strconv.Itoa(config.Port))
	if err == nil {
		log.Fatal(err)
	}
}
