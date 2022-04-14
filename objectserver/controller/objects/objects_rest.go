package objects

import (
	"goodfs/objectserver/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Put(c *gin.Context) {
	fileName := c.Param("name")
	if err := service.Put(fileName, c.Request.Body); err != nil {
		log.Println(err)
		c.Keys["Evict"] = true
		c.Status(http.StatusInternalServerError)
	} else {
		c.Status(http.StatusOK)
	}
}

func Delete(c *gin.Context) {
	name := c.Param("name")
	e := service.Delete(name)
	if e != nil {
		log.Println(e)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Keys["Evict"] = true
	c.Status(http.StatusOK)
}

func Get(c *gin.Context) {
	fileName := c.Param("name")
	e := service.Get(fileName, c.Writer)
	if e != nil {
		log.Println(e)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}
