package objects

import (
	"goodfs/apiserver/model"
	"goodfs/apiserver/model/meta"
	"goodfs/apiserver/service"
	"goodfs/lib/util"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Put(c *gin.Context) {
	req := c.Value("PutReq").(*model.PutReq)
	req.Body = c.Request.Body
	verNum, err := service.StoreObject(req, &meta.MetaDataV2{
		Name: req.Name,
		Versions: []*meta.MetaVersionV2{{
			Size: c.Request.ContentLength,
			Hash: req.Hash,
		}},
	})

	if util.InstanceOf[service.KnownErr](err) {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	} else if err != nil {
		log.Println(err)
		c.Status(http.StatusInternalServerError)
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
		stream, e := service.GetObject(metaVer)
		defer stream.Close()
		if e == service.ErrBadRequest {
			c.AbortWithStatus(http.StatusBadRequest)
		} else if e != nil {
			log.Println(e)
			c.AbortWithStatus(http.StatusServiceUnavailable)
		} else {
			_, e = io.CopyBuffer(c.Writer, stream, make([]byte, 2048))
			if e == nil {
				go service.SendExistingSyncMsg([]byte(metaVer.Hash), model.SyncInsert)
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
