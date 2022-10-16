package temp

import (
	"net/http"
	"objectserver/internal/usecase/pool"

	"github.com/gin-gonic/gin"
)

func FilterExpired(c *gin.Context) {
	id := c.Param("name")
	if !pool.Cache.Has(id) {
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func FilterEmptyRequest(c *gin.Context) {
	if c.Request.ContentLength <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "empty request"})
	}
}
