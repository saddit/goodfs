package controller

import (
	"goodfs/apiserver/global"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func HelpRepairExistFilter(g *gin.Context) {
	g.Header("Count", strconv.FormatUint(uint64(global.ExistFilter.Count()), 10))
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
