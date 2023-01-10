package http

import (
	"common/response"
	"github.com/gin-gonic/gin"
)

type SecurityController struct {
}

func NewSecurityController() *SecurityController {
	return &SecurityController{}
}

func (sc *SecurityController) Register(r gin.IRouter) {
	route := r.Group("security")
	route.GET("/check", sc.Check)
}

func (sc *SecurityController) Check(c *gin.Context) {
	response.Ok(c)
}
