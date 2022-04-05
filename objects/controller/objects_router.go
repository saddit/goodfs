package controller

import (
	"github.com/gin-gonic/gin"
)

func Router(r gin.IRouter) {
	r.GET("/objects/:name", xget)
	r.PUT("/objects/:name", xput)
}
