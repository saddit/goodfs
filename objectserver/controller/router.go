package controller

import (
	"goodfs/objectserver/controller/objects"
	"goodfs/objectserver/controller/temp"

	"github.com/gin-gonic/gin"
)

func Router(r gin.IRouter) {
	r.GET("/objects/:name", objects.GetFromCache, objects.Get, objects.SaveToCache)
	//Deprecated
	r.PUT("/objects/:name", objects.SaveToCache, objects.Put, objects.RemoveCache)
	r.DELETE("/objects/:name", objects.Delete, objects.RemoveCache)

	r.POST("/temp/:name", temp.Post)
	r.PATCH("/temp/:name", temp.FilterExpired, temp.Patch)
	r.DELETE("/temp/:name", temp.FilterExpired, temp.Delete)
	r.HEAD("/temp/:name", temp.FilterExpired, temp.Head)
	r.GET("/temp/:name", temp.FilterExpired, temp.Get)
	r.PUT("/temp/:name", temp.FilterExpired, temp.Put)
}
