package objects

import (
	"goodfs/api/service"
	"goodfs/util"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func put(w http.ResponseWriter, r *http.Request) {
	fileName, ok := util.GetPathVariable(r, 2)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := service.StoreObject(r.Body, fileName)
	if err == service.ErrorServiceUnavailable {
		log.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
	} else if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	fileName, ok := util.GetPathVariable(r, 2)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ip, ok := service.LocateFile(fileName)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	stream, e := service.GetObject(ip, fileName)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
	}
	io.CopyBuffer(w, stream, make([]byte, 2048))
}

func xput(c *gin.Context) {
	fileName := c.Param("name")
	err := service.StoreObject(c.Request.Body, fileName)
	if err == service.ErrorServiceUnavailable {
		c.AbortWithError(http.StatusServiceUnavailable, err)
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	c.Status(http.StatusOK)
}

func xget(c *gin.Context) {
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
