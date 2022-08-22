package request

import (
	"common/util"

	"github.com/gin-gonic/gin"
)

func GetQryInt(key string, c *gin.Context) (int, bool) {
	if v, ok := c.GetQuery(key); ok {
		return util.ToInt(v), true
	}
	return 0, false
}