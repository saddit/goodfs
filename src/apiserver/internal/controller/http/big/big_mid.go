package big

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"common/response"
	"common/util"

	"github.com/gin-gonic/gin"
)

func FilterDuplicates(obj usecase.IObjectService) gin.HandlerFunc {
	return func(g *gin.Context) {
		var req entity.BigPostReq
		if err := req.Bind(g); err != nil {
			response.BadRequestErr(err, g).Abort()
			return
		}
		req.Ext = util.GetFileExtOrDefault(req.Name, false, "bytes")
		g.Set("BigPostReq", &req)
		if ips, ok := obj.LocateObject(req.Hash); ok {
			ver, err := obj.StoreObject(&entity.PutReq{
				Name:   req.Name,
				Hash:   req.Hash,
				Bucket: req.Bucket,
				Ext:    req.Ext,
				Locate: ips,
			}, &entity.Metadata{
				Name: req.Name,
				Versions: []*entity.Version{{
					Size: req.Size,
					Hash: req.Hash,
				}},
			})
			if err != nil {
				response.FailErr(err, g).Abort()
				return
			}
			response.OkJson(&entity.PutResp{
				Name:    req.Name,
				Bucket:  req.Bucket,
				Version: ver,
			}, g).Abort()
		}
	}
}
