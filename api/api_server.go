package api

import (
	"goodfs/api/config"
	"net/http"
	"strconv"
)

func Start() {
	http.HandleFunc("/", nil)
	http.ListenAndServe(":"+strconv.Itoa(config.Port), nil)
}
