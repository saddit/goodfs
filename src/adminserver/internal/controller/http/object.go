package http

import (
	"adminserver/internal/entity"
	"adminserver/internal/usecase/logic"
	"common/response"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
)

type ObjectsController struct {
}

func NewObjectsController() *ObjectsController {
	return &ObjectsController{}
}

func (oc *ObjectsController) Register(route gin.IRouter) {
	route.Group("objects").
		GET("/download/:name").
		PUT("/upload").
		POST("/join/:serverId").
		POST("/leave/:serverId")
}

func (oc *ObjectsController) Upload(c *gin.Context) {
	body := struct {
		File *multipart.FileHeader `form:"file" binding:"required"`
	}{}
	if err := c.ShouldBind(&body); err != nil {
		response.FailErr(err, c)
		return
	}
	if err := logic.NewObjects().Upload(body.File); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Created(c)
}

func (oc *ObjectsController) Download(c *gin.Context) {
	body := struct {
		entity.Binder
		Name    string `uri:"name"`
		Version int    `form:"version"`
	}{}
	if err := body.Bind(c, false); err != nil {
		response.FailErr(err, c)
		return
	}
	reader, err := logic.NewObjects().Download(body.Name, body.Version)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	if _, err = io.Copy(c.Writer, reader); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}

func (oc *ObjectsController) Join(c *gin.Context) {
	if err := logic.NewObjects().JoinCluster(c.Param("serverId")); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}

func (oc *ObjectsController) Leave(c *gin.Context) {
	if err := logic.NewObjects().LeaveCluster(c.Param("serverId")); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}
