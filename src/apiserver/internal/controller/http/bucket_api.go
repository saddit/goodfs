package http

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/repo"
	"common/response"
	"github.com/gin-gonic/gin"
)

type BucketController struct {
	Repo repo.IBucketRepo
}

func NewBucketController(repo repo.IBucketRepo) *BucketController {
	return &BucketController{Repo: repo}
}

func (lc *BucketController) Register(r gin.IRoutes) {
	r.POST("/bucket", lc.Create)
	r.GET("/bucket/:name", lc.Get)
	r.PUT("/bucket/:name", lc.Update)
	r.DELETE("/bucket/:name", lc.Delete)
}

func (lc *BucketController) Create(c *gin.Context) {
	var i entity.Bucket
	if err := c.ShouldBindJSON(&i); err != nil {
		response.FailErr(err, c)
		return
	}
	if i.Name == "" {
		response.BadRequestMsg("name is required", c)
		return
	}
	if err := lc.Repo.Create(&i); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Created(c)
}

func (lc *BucketController) Update(c *gin.Context) {
	var i entity.Bucket
	if err := c.ShouldBindJSON(&i); err != nil {
		response.FailErr(err, c)
		return
	}
	i.Name = c.Param("name")
	if err := lc.Repo.Update(&i); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}

func (lc *BucketController) Delete(c *gin.Context) {
	if err := lc.Repo.Delete(c.Param("name")); err != nil {
		response.FailErr(err, c)
		return
	}
	response.NoContent(c)
}

func (lc *BucketController) Get(c *gin.Context) {
	data, err := lc.Repo.Get(c.Param("name"))
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.OkJson(data, c)
}
