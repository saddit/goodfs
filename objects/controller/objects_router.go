package controller

import (
	"github.com/gin-gonic/gin"
)

func Router(r gin.IRouter) {
	const rest = "/objects/:name"

	r.GET(rest, xget)
	r.PUT(rest, xput)
	r.DELETE(rest, delete)
}
