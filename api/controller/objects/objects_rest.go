package objects

import (
	"goodfs/api/service"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Put(c *gin.Context) {
	fileName := c.Param("name")
	err := service.StoreObject(c.Request.Body, fileName)
	if err == service.ErrorServiceUnavailable {
		c.AbortWithError(http.StatusServiceUnavailable, err)
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.Status(http.StatusOK)
}

func Get(c *gin.Context) {
	fileName := c.Param("name")
	ip, ok := service.LocateFile(fileName)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	stream, e := service.GetObject(ip, fileName)
	if e != nil {
		log.Println(e)
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	io.CopyBuffer(c.Writer, stream, make([]byte, 2048))
	c.Status(http.StatusOK)
}
