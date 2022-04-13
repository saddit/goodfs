package main

import (
	"goodfs/api/config"
	"goodfs/api/controller"
	"goodfs/api/controller/heartbeat"
	"goodfs/api/global"
	"log"
	"strconv"
	"time"

	"github.com/838239178/goodmq"
	"github.com/gin-gonic/gin"
)

func initialize() {
	goodmq.RecoverDelay = 3 * time.Second
	global.AmqpConnection = goodmq.NewAmqpConnection(config.AmqpAddress)
}

func close() {
	global.AmqpConnection.Close()
}

func main() {
	initialize()
	defer close()

	go heartbeat.ListenHeartbeat()

	router := gin.Default()

	api := router.Group("/api")

	controller.Router(api)

	err := router.Run(":" + strconv.Itoa(config.Port))
	if err == nil {
		log.Fatal(err)
	}
}
