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

	if ext, ok := util.GetFileExt(req.Name, true); ok {
		req.FileName = req.Hash
		req.Ext = ext
		if req.Locate, ok = service.LocateFile(req.Hash); !ok {
			req.Locate = nil
		}
	} else {
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "No extension name"})
	}
}
