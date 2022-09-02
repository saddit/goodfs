package request

import (
	"common/util"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ContentTypeJSON = "application/json"
	ContentTypeUrlEncode = "application/x-www-form-urlencoded"
)

func GetQryInt(key string, c *gin.Context) (int, bool) {
	if v, ok := c.GetQuery(key); ok {
		return util.ToInt(v), true
	}
	return 0, false
}

func GetReq(body io.Reader, method, url, contentType string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return req, nil
}

func GetPutReq(body io.Reader, url, contentType string) (*http.Request, error) {
	return GetReq(body, http.MethodPut, url, contentType)
}

func GetDeleteReq(url string) (*http.Request, error) {
	return GetReq(nil, http.MethodDelete, url, ContentTypeUrlEncode)
}