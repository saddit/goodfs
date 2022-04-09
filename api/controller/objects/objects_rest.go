package objects

import (
	"goodfs/api/model"
	"goodfs/api/model/meta"
	"goodfs/api/service"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Put(c *gin.Context) {
	var req model.PutReq
	if err := req.Bind(c); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	verNum, err := service.StoreObject(c.Request.Body, req.Name, &meta.MetaVersion{
		Hash: req.Hash,
		Size: c.Request.ContentLength,
	})

	if err == service.ErrBadRequest {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	} else if err == service.ErrServiceUnavailable {
		c.AbortWithError(http.StatusServiceUnavailable, err)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, model.PutResp{
		Name:    req.Name,
		Version: verNum,
	})
}

func Get(c *gin.Context) {
	var req model.GetReq
	if e := req.Bind(c); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
		return
	}
	
	if metaVer, ok := service.GetMetaVersion(req.Name, req.Version); ok {
		stream, e := service.GetObject(req.Name, metaVer)
		if e == service.ErrBadRequest {
			c.AbortWithStatus(http.StatusBadRequest)
		} else if e != nil {
			log.Println(e)
			c.AbortWithStatus(http.StatusServiceUnavailable)
		} else {
			io.CopyBuffer(c.Writer, stream, make([]byte, 2048))
			c.Status(http.StatusOK)
		}
	} else {
		c.AbortWithStatus(http.StatusNotFound)
	}
}
