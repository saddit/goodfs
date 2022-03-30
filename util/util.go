package util

import (
	"net/http"
	"strings"
)

func GetPathVariable(req *http.Request, no int) (string, bool) {
	splits := strings.Split(req.URL.EscapedPath(), "/")
	if len(splits) <= no+1 {
		return "", false
	}
	return splits[no+1], true
}
