package objects

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"common/datasize"
	"common/logs"
	"common/response"
	"common/util"

	"github.com/gin-gonic/gin"
)

func ValidatePut(obj usecase.IObjectService) gin.HandlerFunc {
	return func(g *gin.Context) {
		var req entity.PutReq
		if err := req.Bind(g); err != nil {
			response.BadRequestErr(err, g).Abort()
			return
		} else if g.Request.ContentLength == 0 {
			response.BadRequestMsg("empty request", g).Abort()
			return
		} else {
			defer g.Set("PutReq", &req)
		}
		if req.Store == 0 {
			if g.Request.ContentLength > int64(datasize.MB*48) {
				req.Store = entity.ECReedSolomon
			} else {
				req.Store = entity.MultiReplication
			}
		}
		req.FileName = req.Hash
		if ext, ok := util.GetFileExt(req.Name, false); ok {
			req.Ext = ext
		} else {
			req.Ext = "bytes"
		}
		if loc, ok := obj.LocateObject(req.Hash); ok {
			logs.Std().Debugf("find locates for %s: %s", req.Hash, loc)
			req.Locate = loc
		}
	}
}
