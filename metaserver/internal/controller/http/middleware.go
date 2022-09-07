package http

import (
	"common/hashslot"
	"common/logs"
	"common/util"
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
		// get slot's location of this key
		location, err := hashslot.GetStringIdentify(name, pool.HashSlots)
		if err != nil {
			logs.Std().Error(err)
			c.Status(http.StatusServiceUnavailable)
			c.Abort()
			return
		}
		// if slot is not in this server, redirect request
		if location != util.GetHostPort(pool.Config.Port) {
			c.Redirect(http.StatusSeeOther, location)
			c.Abort()
			return
		}
	}
	c.Next()
}