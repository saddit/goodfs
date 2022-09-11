package versions

import (
	"common/response"
	"github.com/gin-gonic/gin"
)

func Get(g *gin.Context) {
	response.Ok(g)
}
