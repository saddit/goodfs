package main

import (
	"fmt"
	"goodfs/objects/config"
	"goodfs/objects/controller"
	"goodfs/objects/global"
	"goodfs/objects/heartbeat"
	"goodfs/objects/locate"
	"os"
	"strconv"

	"github.com/838239178/goodmq"
	"github.com/gin-gonic/gin"
)

func initialize() {
	hn, e := os.Hostname()
	if e != nil {
		panic(e)
	}
	config.LocalAddr = fmt.Sprintf("%v:%v", hn, config.Port)
	global.AmqpConnection = goodmq.NewAmqpConnection(config.AmqpAddress)
}

func close() {
	global.AmqpConnection.Close()
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
