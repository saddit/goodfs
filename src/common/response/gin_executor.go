package response

import (
	"common/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GinExecutor struct {
	ctx *gin.Context
}

func Exec(c *gin.Context) *GinExecutor {
	return &GinExecutor{c}
}

func (ge *GinExecutor) Abort() *GinExecutor {
	ge.ctx.Abort()
	return ge
}

func (ge *GinExecutor) Status(code int) *GinExecutor {
	ge.ctx.Status(code)
	return ge
}

func (ge *GinExecutor) Fail(code int, msg string) {
	ge.ctx.JSON(code, &FailureResp{
		Success: false,
		Message: msg,
	})
}

func (ge *GinExecutor) JSON(body any) *GinExecutor {
	ge.ctx.JSON(http.StatusOK, body)
	return ge
}

func (ge *GinExecutor) Header(hd gin.H) *GinExecutor {
	for k, v := range hd {
		ge.ctx.Header(k, util.ToString(v))
	}
	return ge
}

func (ge *GinExecutor) Redirect(code int, s string) *GinExecutor {
	ge.ctx.Redirect(code, s)
	return ge
}
