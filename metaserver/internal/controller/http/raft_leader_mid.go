package http

import (
	"metaserver/internal/usecase/pool"
	"net/http"

	"github.com/gin-gonic/gin"
)

func isWriteMethod(method string) bool {
	return method == http.MethodPut ||
		method == http.MethodDelete ||
		method == http.MethodPatch ||
		method == http.MethodPost 
}

func CheckLeaderInRaftMode(c *gin.Context) {
	if pool.RaftWrapper.Enabled {
		if isWriteMethod(c.Request.Method) && !pool.RaftWrapper.IsLeader() {
			c.Status(http.StatusServiceUnavailable)
			c.Abort()
			return
		}
	}
	c.Next()
}