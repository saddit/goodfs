package request

import (
	"bytes"
	"common/util"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

const (
	ContentTypeJSON      = "application/json"
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

func JsonReq(method string, url string, data any) (*http.Request, error) {
	bt, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return GetReq(bytes.NewBuffer(bt), method, url, ContentTypeJSON)
}

func UrlValuesEncode(url string, form *url.Values) (*http.Request, error) {
	return http.NewRequest(http.MethodGet, fmt.Sprint(url, "?", form.Encode()), nil)
}
