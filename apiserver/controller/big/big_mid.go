package big

import (
	"github.com/gin-gonic/gin"
	"goodfs/apiserver/model"
	"goodfs/apiserver/model/meta"
	"goodfs/apiserver/service"
	"goodfs/lib/util"
	"log"
	"net/http"
)

func FilterDuplicates(g *gin.Context) {
	var req model.BigPostReq
	if e := req.Bind(g); e != nil {
		g.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": e.Error()})
		return
	} else {
		g.Set("BigPostReq", &req)
	}
	if ips, ok := service.LocateFile(req.Hash); ok {
		ver, e := service.StoreObject(&model.PutReq{
			Name:     req.Name,
			Hash:     req.Hash,
			Ext:      "bytes",
			Locate:   ips,
			FileName: req.Hash,
		}, &meta.Data{
			Name: req.Name,
			Versions: []*meta.Version{{
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
		g.AbortWithStatusJSON(http.StatusOK, model.PutResp{
			Name:    req.Name,
			Version: ver,
		})
	}
}
