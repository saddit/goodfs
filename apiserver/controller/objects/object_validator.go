package objects

import (
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
		req.FileName = req.Hash + ext
		if req.Locate, ok = service.LocateFile(req.FileName); !ok {
			req.Body = g.Request.Body
		}
	} else {
		g.AbortWithStatus(http.StatusBadRequest)
	}
}
