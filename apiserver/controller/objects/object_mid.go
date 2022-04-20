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
	} else if g.Request.Body == nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "Empty request"})
		return
	} else {
		defer g.Set("PutReq", &req)
	}

	if ext, ok := util.GetFileExt(req.Name, true); ok {
		req.FileName = req.Hash
		req.Ext = ext
		//用过滤器进行第一级筛查，表示可能存在才进行Locate的获取
		if global.ExistFilter.Lookup([]byte(req.FileName)) {
			if req.Locate, ok = service.LocateFile(req.FileName); !ok {
				go service.SendExistingSyncMsg([]byte(req.FileName), model.SyncDelete)
			}
		}
	} else {
		g.AbortWithStatus(http.StatusBadRequest)
	}
}

func ChangeExisting(g *gin.Context) {
	if g.Writer.Status() == http.StatusOK {
		if req, ok := g.Get("PutReq"); ok {
			//Put成功，更新过滤器
			key := []byte(req.(*model.PutReq).FileName)
			go service.SendExistingSyncMsg(key, model.SyncInsert)
		}
	}
}
