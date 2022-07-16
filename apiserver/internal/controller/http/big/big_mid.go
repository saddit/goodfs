package big

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/service"
	"common/util"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func FilterDuplicates(g *gin.Context) {
	var req entity.BigPostReq
	if e := req.Bind(g); e != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
		return
	} else {
		g.Set("BigPostReq", &req)
	}
	if ips, ok := ObjectService.LocateObject(req.Hash); ok {
		ver, e := ObjectService.StoreObject(&entity.PutReq{
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
		if util.InstanceOf[service.KnownErr](e) {
			g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
			return
		} else if e != nil {
			log.Println(e)
			g.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		g.AbortWithStatusJSON(http.StatusOK, entity.PutResp{
			Name:    req.Name,
			Version: ver,
		})
	}
}
