package http

import (
	"common/response"
	"metaserver/internal/usecase/logic"
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

func CheckKeySlot(c *gin.Context) {
	if isWriteMethod(c.Request.Method) {
		name := c.Param("name")
		if name == "" {
			c.Next()
			return
		}
		if ok, other := logic.NewHashSlot().IsKeyOnThisServer(name); !ok {
			response.Exec(c).Redirect(http.StatusSeeOther, other)
			return
		}
	}
	c.Next()
}