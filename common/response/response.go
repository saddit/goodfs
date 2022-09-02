package response

import (
	"common/logs"
	"common/util"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type FailureResp struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	SubMessage string `json:"sub_message"`
}

func Ok(c *gin.Context) {
	c.JSON(http.StatusOK, &FailureResp{
		Success: true,
		Message: "success",
	})
}

func OkHeader(h gin.H, c *gin.Context) {
	for k, v := range h {
		c.Header(k, util.ToString(v))
	}
	Ok(c)
}

func OkJson(data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, data)
}

func Created(c *gin.Context) {
	c.Status(http.StatusCreated)
}

func CreatedJson(data interface{}, c *gin.Context) {
	c.JSON(http.StatusCreated, data)
}

func CreatedHeader(h gin.H, c *gin.Context) {
	for k, v := range h {
		c.Header(k, util.ToString(v))
	}
	Created(c)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func NotFound(msg string, c *gin.Context) {
	c.Status(http.StatusNotFound)
}

func NotFoundMsg(msg string, c *gin.Context) {
	c.JSON(http.StatusNotFound, &FailureResp{
		Message: msg,
	})
}

func NotFoundErr(err error, c *gin.Context) {
	c.JSON(http.StatusNotFound, &FailureResp{
		Message: err.Error(),
	})
}

func BadRequestErr(err error, c *gin.Context) {
	c.JSON(http.StatusBadRequest, &FailureResp{
		Message: err.Error(),
	})
}

func BadRequestMsg(msg string, c *gin.Context) {
	c.JSON(http.StatusBadRequest, &FailureResp{
		Message: msg,
	})
}

func FailErr(err error, c *gin.Context) {
	switch err := err.(type) {
	case validator.ValidationErrors, *validator.ValidationErrors:
		BadRequestErr(err, c)
	case *ResponseErr:
		c.JSON(err.Status, &FailureResp{
			Message:    err.Message,
			SubMessage: err.Error(),
		})
	case ResponseErr:
		c.JSON(err.Status, &FailureResp{
			Message:    err.Message,
			SubMessage: err.Error(),
		})
	default:
		logs.Std().Error(fmt.Sprintf("request(%s %s): %+v", c.Request.Method, c.FullPath(), err))
		c.JSON(http.StatusInternalServerError, &FailureResp{
			Message: "系统错误",
		})
		c.Abort()
	}
}
