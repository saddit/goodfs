package controller

import (
	"github.com/gin-gonic/gin"
	"goodfs/apiserver/global"
	"goodfs/lib/util"
	"log"
	"net/http"
)

func HelpRepairExistFilter(g *gin.Context) {
	g.Header("Count", util.NumToString(global.ExistFilter.Count()))
	if err := global.ExistFilter.EncodeBuckets(g.Writer); err != nil {
		log.Println(err)
		g.AbortWithStatus(http.StatusInternalServerError)
	} else {
		g.Status(http.StatusOK)
	}
}

func HelperRouter(r gin.IRouter) {
	r.GET("/exist_filter", HelpRepairExistFilter)
}
