package locate

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Get(ctx *gin.Context) {
	name := ctx.Param("name")
	info, exist := ObjectService.LocateObject(name)
	if !exist {
		ctx.Status(http.StatusNotFound)
	} else {
		ctx.JSON(http.StatusOK, info)
	}
}
