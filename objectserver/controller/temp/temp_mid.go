package temp

import (
	"github.com/gin-gonic/gin"
	"goodfs/objectserver/global"
	"net/http"
)

func FilterExpired(c *gin.Context) {
	id := c.Param("name")
	if !global.Cache.Has(id) {
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func FilterEmptyRequest(c *gin.Context) {
	if c.Request.ContentLength <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "empty request"})
	}
}
