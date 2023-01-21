package http

import (
	"adminserver/internal/entity"
	"adminserver/internal/usecase/logic"
	"adminserver/internal/usecase/pool"
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
	route.Group("objects").Use(RequireToken).
		GET("/download/:name", oc.Download).
		PUT("/upload", oc.Upload).
		POST("/join/:serverId", oc.Join).
		POST("/leave/:serverId", oc.Leave).
		GET("/config/:serverId", oc.GetConfig)
}

func (oc *ObjectsController) Upload(c *gin.Context) {
	body := struct {
		File   *multipart.FileHeader `form:"file" binding:"required"`
		Bucket string                `form:"bucket" binding:"required"`
	}{}
	if err := c.ShouldBind(&body); err != nil {
		response.FailErr(err, c)
		return
	}
	if err := logic.NewObjects().Upload(body.File, body.Bucket, GetAuthToken(c)); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Created(c)
}

func (oc *ObjectsController) Download(c *gin.Context) {
	body := struct {
		Name    string `uri:"name"`
		Bucket  string `form:"bucket"`
		Version int    `form:"version"`
	}{}
	if err := entity.Bind(c, &body, false); err != nil {
		response.FailErr(err, c)
		return
	}
	reader, err := logic.NewObjects().Download(body.Name, body.Bucket, body.Version, GetAuthToken(c))
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

func (oc *ObjectsController) GetConfig(c *gin.Context) {
	sid := c.Param("serverId")
	ip, ok := pool.Discovery.GetService(pool.Config.Discovery.DataServName, sid, true)
	if !ok {
		response.BadRequestMsg("unknown serverId", c)
		return
	}
	jsonData, err := logic.NewObjects().GetConfig(ip)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	if _, err = c.Writer.Write(jsonData); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}
