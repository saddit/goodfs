package response

import (
	"common/logs"
	"common/util"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type FailureResp struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	SubMessage string `json:"sub_message"`
}

func Ok(c *gin.Context) *GinExecutor {
	c.Status(http.StatusOK)
	return &GinExecutor{c}
}

func OkHeader(h gin.H, c *gin.Context) *GinExecutor {
	for k, v := range h {
		c.Header(k, util.ToString(v))
	}
	Ok(c)
	return &GinExecutor{c}
}

func OkJson(data interface{}, c *gin.Context) *GinExecutor {
	c.JSON(http.StatusOK, data)
	return &GinExecutor{c}
}

func Created(c *gin.Context) *GinExecutor {
	c.Status(http.StatusCreated)
	return &GinExecutor{c}
}

func CreatedJson(data interface{}, c *gin.Context) *GinExecutor {
	c.JSON(http.StatusCreated, data)
	return &GinExecutor{c}
}

func CreatedHeader(h gin.H, c *gin.Context) *GinExecutor {
	for k, v := range h {
		c.Header(k, util.ToString(v))
	}
	Created(c)
	return &GinExecutor{c}
}

func NoContent(c *gin.Context) *GinExecutor {
	c.Status(http.StatusNoContent)
	return &GinExecutor{c}
}

func NotFound(c *gin.Context) *GinExecutor {
	c.Status(http.StatusNotFound)
	return &GinExecutor{c}
}

func NotFoundMsg(msg string, c *gin.Context) *GinExecutor {
	c.JSON(http.StatusNotFound, &FailureResp{
		Message: msg,
	})
	return &GinExecutor{c}
}

func NotFoundErr(err error, c *gin.Context) *GinExecutor {
	c.JSON(http.StatusNotFound, &FailureResp{
		Message:    err.Error(),
		SubMessage: "resource doesn't exist",
	})
	return &GinExecutor{c}
}

func BadRequestErr(err error, c *gin.Context) *GinExecutor {
	c.JSON(http.StatusBadRequest, &FailureResp{
		Message:    err.Error(),
		SubMessage: "check parameters",
	})
	return &GinExecutor{c}
}

func BadRequestMsg(msg string, c *gin.Context) *GinExecutor {
	c.JSON(http.StatusBadRequest, &FailureResp{
		Message: msg,
	})
	return &GinExecutor{c}
}

func ServiceUnavailableMsg(msg string, c *gin.Context) *GinExecutor {
	ge := &GinExecutor{c}
	return ge.Status(http.StatusServiceUnavailable).
		JSON(&FailureResp{
			Message: msg,
		})
}

func ServiceUnavailableErr(err error, c *gin.Context) *GinExecutor {
	return ServiceUnavailableMsg(err.Error(), c)
}

func FailErr(err error, c *gin.Context) *GinExecutor {
	switch err := err.(type) {
	case validator.ValidationErrors, *validator.ValidationErrors:
		BadRequestErr(err, c)
	case IErr:
		if IsOk(err.GetStatus()) {
			c.Status(err.GetStatus())
		} else if IsInternal(err.GetStatus()) {
			logs.Std().Errorf("request(%s %s): [%T] %s", c.Request.Method, c.FullPath(), err, err)
			c.JSON(err.GetStatus(), &FailureResp{
				Message:    "system error",
				SubMessage: fmt.Sprintf("%T", err),
			})
		} else {
			c.JSON(err.GetStatus(), &FailureResp{
				Message:    err.Error(),
				SubMessage: err.GetSubMessage(),
			})
		}
	default:
		logs.Std().Errorf("request(%s %s): [%T] %s", c.Request.Method, c.FullPath(), err, err)
		c.JSON(http.StatusInternalServerError, &FailureResp{
			Message:    "system error",
			SubMessage: fmt.Sprintf("%T", err),
		})
		c.Abort()
	}
	return &GinExecutor{c}
}

// TODO: add more mapping
var httpGrpcStatus = map[int]codes.Code{
	http.StatusOK:           codes.OK,
	http.StatusBadRequest:   codes.InvalidArgument,
	http.StatusUnauthorized: codes.Unauthenticated,
	http.StatusForbidden:    codes.PermissionDenied,
	http.StatusNotFound:     codes.NotFound,
}

func GRPCError(err error) error {
	switch err := err.(type) {
	case validator.ValidationErrors, *validator.ValidationErrors:
		return status.Error(codes.InvalidArgument, err.Error())
	case IErr:
		if IsOk(err.GetStatus()) {
			return nil
		} else if IsInternal(err.GetStatus()) {
			logs.Std().Errorf("grpc internal err: %s", err)
			return status.Error(codes.Internal, err.Error())
		} else {
			code, ok := httpGrpcStatus[err.GetStatus()]
			if !ok {
				logs.Std().Errorf("grpc unknown err: %s", err)
				code = codes.Unknown
			}
			return status.Error(code, err.Error())
		}
	default:
		logs.Std().Errorf("grpc internal err: %s", err)
		return status.Error(codes.Internal, err.Error())
	}
}
