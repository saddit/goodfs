package main

import (
	"goodfs/api/config"
	"goodfs/api/controller"
	"goodfs/api/controller/heartbeat"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	go heartbeat.ListenHeartbeat()

	router := gin.Default()

	api := router.Group("/api")

	controller.Router(api)

	err := router.Run(":" + strconv.Itoa(config.Port))
	if err == nil {
		log.Fatal(err)
	}
}
