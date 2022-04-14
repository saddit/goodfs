package controller

import (
	"goodfs/objectserver/controller/objects"
	"goodfs/objectserver/controller/temp"

	"github.com/gin-gonic/gin"
)

func Router(r gin.IRouter) {
	r.GET("/objects/:name", objects.GetFromCache, objects.Get, objects.SaveToCache)
	r.PUT("/objects/:name", objects.SaveToCache, objects.Put, objects.RemoveCache)
	r.DELETE("/objects/:name", objects.Delete, objects.RemoveCache)

	r.POST("/temp/:name", temp.Post)
	r.PATCH("/temp/:name", temp.Patch)
	r.DELETE("/temp/:name", temp.Delete)
	r.PUT("/temp/:name", temp.Put)
}
