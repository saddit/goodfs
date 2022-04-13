package controller

import (
	"goodfs/objectserver/controller/objects"
	"goodfs/objectserver/controller/temp"

	"github.com/gin-gonic/gin"
)

func Router(r gin.IRouter) {
	r.GET("/objects/:name", objects.Get)
	r.PUT("/objects/:name", objects.Put)
	r.DELETE("/objects/:name", objects.Delete)

	r.GET("/temp/:name", temp.Get)
	r.POST("/temp/:name", temp.Post)
	r.PATCH("/temp/:name", temp.Patch)
	r.DELETE("/temp/:name", temp.Delete)
	r.PUT("/temp/:name", temp.Put)
}
