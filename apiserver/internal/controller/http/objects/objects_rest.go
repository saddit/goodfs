package objects

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/service"
	"common/util"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Put(c *gin.Context) {
	req := c.Value("PutReq").(*entity.PutReq)
	req.Body = c.Request.Body
	verNum, err := ObjectService.StoreObject(req, &entity.MetaData{
		Name: req.Name,
		Versions: []*entity.Version{{
			Size: c.Request.ContentLength,
			Hash: req.Hash,
		}},
	})

	if util.InstanceOf[service.KnownErr](err) {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	} else if err != nil {
		AbortInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity.PutResp{
		Name:    req.Name,
		Version: verNum,
	})
}

func Get(c *gin.Context) {
	var req entity.GetReq
	if e := req.Bind(c); e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
		return
	}

	if metaData, ok := MetaService.GetMetadata(req.Name, req.Version); ok {
		metaVer := metaData.Versions[0]
		//表示此版本已经移被删除
		if metaVer.Size == 0 || metaVer.Hash == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		stream, e := ObjectService.GetObject(metaVer)
		defer stream.Close()
		if e == service.ErrBadRequest {
			c.AbortWithStatus(http.StatusBadRequest)
		} else if e != nil {
			AbortServiceUnavailableError(c, e)
		} else {
			if tp, ok := req.Range.Get(); ok {
				if _, e = stream.Seek(tp.First, io.SeekCurrent); e != nil {
					AbortInternalError(c, e)
					return
				}
			}
			_, e = io.CopyBuffer(c.Writer, stream, make([]byte, 2048))
			if e == nil {
				c.Status(http.StatusOK)
			} else {
				AbortInternalError(c, e)
			}
		}
	} else {
		c.AbortWithStatus(http.StatusNotFound)
	}
}
