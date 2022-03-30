package api

import (
	"goodfs/api/config"
	"goodfs/api/heartbeat"
	"goodfs/api/locate"
	"goodfs/api/objects"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Start() {
	go heartbeat.ListenHearbeat()

	router := gin.Default()
	api := router.Group("/api")
	objects.Router(api)
	locate.Router(api)

	router.Run(":" + strconv.Itoa(config.Port))

	// http.HandleFunc("/api/objects/", objects.Handler)
	// http.HandleFunc("/api/locate/", locate.Handler)
}
