package big

import (
	"apiserver/internal/entity"
	"common/response"
	"common/util"
	"github.com/gin-gonic/gin"
)

func FilterDuplicates(g *gin.Context) {
	var req entity.BigPostReq
	if err := req.Bind(g); err != nil {
		response.BadRequestErr(err, g).Abort()
		return
	} else {
		g.Set("BigPostReq", &req)
	}
	if ips, ok := ObjectService.LocateObject(req.Hash); ok {
		ver, err := ObjectService.StoreObject(&entity.PutReq{
			Name:     req.Name,
			Hash:     req.Hash,
			Ext:      util.GetFileExtOrDefault(req.Name, true, "bytes"),
			Locate:   ips,
			FileName: req.Hash,
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
			Version: ver,
		}, g).Abort()
	}
}
