package controller

import (
	"goodfs/objectserver/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func xput(c *gin.Context) {
	fileName := c.Param("name")
	err := service.Put(fileName, c.Request.Body)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func delete(c *gin.Context) {
	name := c.Param("name")
	e := service.Delete(name)
	if e != nil {
		log.Println(e)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}

func xget(c *gin.Context) {
	fileName := c.Param("name")
	e := service.Get(fileName, c.Writer)
	if e != nil {
		log.Println(e)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}
