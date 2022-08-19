package response

import (
	"common/logs"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type FailureResp struct {
	Success    bool
	Message    string `json:"message"`
	SubMessage string `json:"sub_message"`
}

func Ok(c *gin.Context) {
	c.JSON(200, &FailureResp{
		Success: true,
		Message: "success",
	})
}

func OkJson(data interface{}, c *gin.Context) {
	c.JSON(200, data)
}

func Created(c *gin.Context) {
	c.Status(201)
}

func CreatedJson(data interface{}, c *gin.Context) {
	c.JSON(201, data)
}

func NotFound(msg string, c *gin.Context) {
	c.Status(404)
}

func NotFoundMsg(msg string, c *gin.Context) {
	c.JSON(404, &FailureResp{
		Message: msg,
	})
}

func NotFoundErr(err error, c *gin.Context) {
	c.JSON(404, &FailureResp{
		Message: err.Error(),
	})
}

func BadRequestErr(err error, c *gin.Context) {
	c.JSON(400, &FailureResp{
		Message: err.Error(),
	})
}

func BadRequestMsg(msg string, c *gin.Context) {
	c.JSON(400, &FailureResp{
		Message: msg,
	})
}

func FailErr(err error, c *gin.Context) {
	var respErr ResponseErr
	var validErr validator.ValidationErrors
	if errors.As(err, &respErr) {
		c.JSON(respErr.Status, &FailureResp{
			Message: respErr.Message,
			SubMessage: err.Error(),
		})
	} else if errors.As(err, &validErr) {
		BadRequestErr(validErr, c)
	} else {
		logs.Std().Error(c.Error(err))
		c.JSON(500, &FailureResp{
			Message: "系统错误",
		})
		c.Abort()
	}
}
