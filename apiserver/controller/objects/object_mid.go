package objects

import (
	"goodfs/apiserver/model"
	"goodfs/apiserver/service"
	"goodfs/lib/util"
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

	req.FileName = req.Hash
	if ext, ok := util.GetFileExt(req.Name, false); ok {
		req.Ext = ext
	} else {
		req.Ext = "bytes"
	}
	if loc, ok := service.LocateFile(req.Hash); ok {
		req.Locate = loc
	}
}
