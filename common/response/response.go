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
	c.Status(http.StatusOK)
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
		SubMessage: "resource doesn't exist",
	})
}

func BadRequestErr(err error, c *gin.Context) {
	c.JSON(http.StatusBadRequest, &FailureResp{
		Message: err.Error(),
		SubMessage: "check parameters",
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
	case IResponseErr:
		if IsOk(err.GetStatus()) {
			c.Status(err.GetStatus())
		} else {
			c.JSON(err.GetStatus(), &FailureResp{
				Message:    err.GetMessage(),
				SubMessage: err.Error(),
			})
		}
	default:
		logs.Std().Error(fmt.Sprintf("request(%s %s): [%T] %s", c.Request.Method, c.FullPath(), err, err))
		c.JSON(http.StatusInternalServerError, &FailureResp{
			Message: "system error",
			SubMessage: fmt.Sprintf("%T", err),
		})
		c.Abort()
	}
}
