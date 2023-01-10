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
		GET("/download/:name", oc.Download).
		PUT("/upload", oc.Upload).
		POST("/join/:serverId", oc.Join).
		POST("/leave/:serverId", oc.Leave)
}

func (oc *ObjectsController) Upload(c *gin.Context) {
	body := struct {
		File *multipart.FileHeader `form:"file" binding:"required"`
	}{}
	if err := c.ShouldBind(&body); err != nil {
		response.FailErr(err, c)
		return
	}
	if err := logic.NewObjects().Upload(body.File, GetAuthToken(c)); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Created(c)
}

func (oc *ObjectsController) Download(c *gin.Context) {
	body := struct {
		Name    string `uri:"name"`
		Version int    `form:"version"`
	}{}
	if err := entity.Bind(c, &body, false); err != nil {
		response.FailErr(err, c)
		return
	}
	reader, err := logic.NewObjects().Download(body.Name, body.Version, GetAuthToken(c))
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
