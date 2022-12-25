package objects

import (
	"log"
	"net/http"
	"objectserver/internal/usecase/service"

	"github.com/gin-gonic/gin"
)

func Put(c *gin.Context) {
	fileName := c.Param("name")
	if err := service.Put(fileName, c.Request.Body); err != nil {
		log.Println(err)
		c.Set("Evict", true)
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
	c.Set("Evict", true)
	c.Status(http.StatusNoContent)
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

func Head(c *gin.Context) {
	fileName := c.Param("name")
	if ok := service.Exist(fileName); ok {
		c.Status(http.StatusOK)
		return
	}
	c.Status(http.StatusNotFound)
}
