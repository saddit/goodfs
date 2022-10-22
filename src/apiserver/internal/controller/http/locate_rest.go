package http

import (
	"apiserver/internal/usecase"
	"common/response"

	"github.com/gin-gonic/gin"
)

type LocateController struct {
	objectService usecase.IObjectService
}

func NewLocateController(obj usecase.IObjectService) *LocateController {
	return &LocateController{obj}
}

func (lc *LocateController) Register(r gin.IRoutes) {
	r.GET("/locate/:name", lc.Get)
}

func (lc *LocateController) Get(c *gin.Context) {
	name := c.Param("name")
	info, exist := lc.objectService.LocateObject(name)
	if !exist {
		response.NotFound(c)
	} else {
		response.Ok(c).JSON(info)
	}
}
