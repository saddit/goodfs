package objects

import (
	"goodfs/objects/service"
	"goodfs/util"
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
	err := service.Put(fileName, r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	fileName, ok := util.GetPathVariable(r, 2)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
	}
	e := service.Get(fileName, w)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusNotFound)
	}
}

func xput(c *gin.Context) {
	fileName := c.Param("name")
	err := service.Put(fileName, c.Request.Body)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.Status(http.StatusOK)
}

func xget(c *gin.Context) {
	fileName := c.Param("name")
	e := service.Get(fileName, c.Writer)
	if e != nil {
		log.Println(e)
		c.AbortWithStatus(http.StatusNotFound)
	}
	c.Status(http.StatusOK)
}
