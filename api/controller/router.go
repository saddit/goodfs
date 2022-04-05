package controller

import (
	"goodfs/api/controller/locate"
	"goodfs/api/controller/objects"

	"github.com/gin-gonic/gin"
)

func Router(r gin.IRouter) {
	r.PUT("/objects/:name", objects.Put)
	r.GET("/objects/:name", objects.Get)
	r.GET("/locate/:name", locate.Get)
}
