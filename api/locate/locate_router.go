package locate

import (
	"encoding/json"
	"goodfs/api/service"
	"goodfs/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Handler(w http.ResponseWriter, req *http.Request) {
	m := req.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	name, ok := util.GetPathVariable(req, 2)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	info, exist := service.LocateFile(name)
	if !exist {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	b, _ := json.Marshal(info)
	w.Write(b)
}

func get(ctx *gin.Context) {
	name := ctx.Param("name")
	info, exist := service.LocateFile(name)
	if !exist {
		ctx.AbortWithStatus(http.StatusNotFound)
	} else {
		ctx.JSON(http.StatusOK, info)
	}
}

func Router(r gin.IRouter) {
	r.GET("/locate/:name", get)
}
