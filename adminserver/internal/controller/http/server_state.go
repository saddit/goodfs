package http

import "github.com/gin-gonic/gin"

type ServerStateController struct {
}

func NewServerStateController() *ServerStateController {
	return &ServerStateController{}
}

func (ss *ServerStateController) Register(r gin.IRoutes) {

}
