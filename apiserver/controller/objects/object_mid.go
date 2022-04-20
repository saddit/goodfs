package objects

import (
	"goodfs/apiserver/global"
	"goodfs/apiserver/model"
	"goodfs/apiserver/service"
	"goodfs/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ValidatePut(g *gin.Context) {
	var req model.PutReq
	if err := req.Bind(g); err != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	} else {
		g.Keys["PutReq"] = &req
	}

	if ext, ok := util.GetFileExt(req.Name, true); ok {
		req.FileName = req.Hash
		req.Ext = ext
		//用过滤器进行第一级筛查，表示可能存在才进行Locate的获取
		if global.ExistFilter.Lookup([]byte(req.FileName)) {
			if req.Locate, ok = service.LocateFile(req.FileName); !ok {
				req.Body = g.Request.Body
			}
		}
	} else {
		g.AbortWithStatus(http.StatusBadRequest)
	}
}

func ChangeExisting(g *gin.Context) {
	if g.Request.Response.StatusCode == http.StatusOK {
		if req, ok := g.Keys["PutReq"].(*model.PutReq); ok {
			//Put成功，更新过滤器
			key := []byte(req.FileName)
			_ = service.SendExistingSyncMsg(key, model.SyncInsert)
		}
	}
}
