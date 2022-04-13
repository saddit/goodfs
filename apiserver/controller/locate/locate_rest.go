package locate

import (
	"goodfs/apiserver/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Get(ctx *gin.Context) {
	name := ctx.Param("name")
	info, exist := service.LocateFile(name)
	if !exist {
		ctx.AbortWithStatus(http.StatusNotFound)
	} else {
		ctx.JSON(http.StatusOK, info)
	}
}
