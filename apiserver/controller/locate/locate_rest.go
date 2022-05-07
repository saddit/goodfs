package locate

import (
	"github.com/gin-gonic/gin"
	"goodfs/apiserver/service"
	"net/http"
)

func Get(ctx *gin.Context) {
	name := ctx.Param("name")
	info, exist := service.LocateFile(name)
	if !exist {
		ctx.Status(http.StatusNotFound)
	} else {
		ctx.JSON(http.StatusOK, info)
	}
}
