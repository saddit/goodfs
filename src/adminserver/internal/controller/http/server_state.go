package http

import (
	"github.com/gin-gonic/gin"
)

type ServerStateController struct {
}

func NewServerStateController() *ServerStateController {
	return &ServerStateController{}
}

func (ss *ServerStateController) Register(r gin.IRouter) {
	r.Group("server").GET("/register_info")
}

func (ss *ServerStateController) RegisterInfo(c *gin.Context) {
	// TODO
}
