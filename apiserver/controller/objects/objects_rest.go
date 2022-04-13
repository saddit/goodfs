package objects

import (
	"goodfs/apiserver/model"
	"goodfs/apiserver/model/meta"
	"goodfs/apiserver/service"
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
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else if err == service.ErrServiceUnavailable {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	} else if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
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

	if metaData, ok := service.GetMetaData(req.Name, req.Version); ok {
		metaVer := metaData.Versions[0]
		//表示此版本已经移被删除
		if metaVer.Size == 0 || metaVer.Hash == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		stream, e := service.GetObject(req.Name, metaVer)
		if e == service.ErrBadRequest {
			c.AbortWithStatus(http.StatusBadRequest)
		} else if e != nil {
			log.Println(e)
			c.AbortWithStatus(http.StatusServiceUnavailable)
		} else {
			_, e = io.CopyBuffer(c.Writer, stream, make([]byte, 2048))
			if e == nil {
				c.Status(http.StatusOK)
			} else {
				log.Println(e)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}
	} else {
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func Delete(c *gin.Context) {

}
