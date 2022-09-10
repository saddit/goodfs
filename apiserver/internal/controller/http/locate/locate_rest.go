package locate

import (
	"common/response"
	"github.com/gin-gonic/gin"
)

func Get(c *gin.Context) {
	name := c.Param("name")
	info, exist := ObjectService.LocateObject(name)
	if !exist {
		response.NotFound(c)
	} else {
		response.Ok(c).JSON(info)
	}
}
