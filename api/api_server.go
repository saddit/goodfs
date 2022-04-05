package main

import (
	"goodfs/api/config"
	"goodfs/api/controller"
	"goodfs/api/controller/heartbeat"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	go heartbeat.ListenHearbeat()

	router := gin.Default()

	api := router.Group("/api")

	controller.Router(api)

	router.Run(":" + strconv.Itoa(config.Port))
}
