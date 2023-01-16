package http

import (
	"common/response"
	"github.com/gin-gonic/gin"
	"metaserver/internal/entity"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/logic"
	"net/http"
)

type BucketController struct {
	service usecase.BucketService
}

func NewBucketController(service usecase.BucketService) *BucketController {
	return &BucketController{service: service}
}

func (b *BucketController) RegisterRoute(route gin.IRouter) {
	rt := route.Group("bucket")
	rt.POST("/", b.CreateNew)
	rt.POST("", b.CreateNew)
	rt.DELETE("/:name", b.Delete)
	rt.GET("/:name", b.Get)
	rt.GET("/list", b.List)
	rt.PUT("/:name", b.Update)
}

func (b *BucketController) Get(c *gin.Context) {
	name := c.Param("name")
	data, err := b.service.Get(name)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.OkJson(data, c)
}

func (b *BucketController) CreateNew(c *gin.Context) {
	var data entity.Bucket
	if err := c.ShouldBindJSON(&data); err != nil {
		response.FailErr(err, c)
		return
	}
	if data.Name == "" {
		response.BadRequestMsg("name required", c)
		return
	}
	// post request has not 'name' param. need check again here
	if ok, other := logic.NewHashSlot().IsKeyOnThisServer(data.Name); !ok {
		response.Exec(c).Redirect(http.StatusSeeOther, other)
		return
	}
	if err := b.service.Create(&data); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Created(c)
}

func (b *BucketController) Update(c *gin.Context) {
	name := c.Param("name")
	var data entity.Bucket
	if err := c.ShouldBindJSON(&data); err != nil {
		response.FailErr(err, c)
		return
	}
	data.Name = name
	if err := b.service.Update(&data); err != nil {
		response.FailErr(err, c)
		return
	}
}

func (b *BucketController) Delete(c *gin.Context) {
	name := c.Param("name")
	err := b.service.Remove(name)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.NoContent(c)
}

func (b *BucketController) List(c *gin.Context) {
	req := &struct {
		Prefix   string `form:"prefix"`
		PageSize int    `form:"page_size" binding:"required,lte=10000"`
	}{}
	if err := c.ShouldBindQuery(req); err != nil {
		response.FailErr(err, c)
		return
	}
	res, total, err := b.service.List(req.Prefix, req.PageSize)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.Exec(c).Header(gin.H{
		"X-Total-Count": total,
	}).JSON(res)
}
