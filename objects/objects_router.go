package objects

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m == http.MethodPut {
		put(w, r)
	} else if m == http.MethodGet {
		get(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func Router(r gin.IRouter) {
	r.GET("/objects/:name", xget)
	r.PUT("/objects/:name", xput)
}
