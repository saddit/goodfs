package auth

import (
	"apiserver/internal/usecase/componet/auth/credential"
	"bytes"
	"common/response"
	"common/util"
	"errors"
	"net/http"
	"net/url"
	"apiserver/config"

	"github.com/gin-gonic/gin"
)

type CallbackValidator struct {
	cfg *config.CallbackConfig
	cli *http.Client
}

func NewCallbackValidator(cli *http.Client, cfg *config.CallbackConfig) *CallbackValidator {
	return &CallbackValidator{cfg: cfg, cli: cli}
}

func (cv *CallbackValidator) Verify(token Credential) error {
	if !cv.cfg.Enable {
		return errors.New("not enable password verification")
	}
	body := bytes.NewBuffer([]byte(token.GetUsername()))
	uri, err := url.Parse(cv.cfg.Url)
	if err != nil {
		return err
	}
	uri.RawQuery = url.Values(token.GetExtra()).Encode()
	resp, err := cv.cli.Post(uri.RequestURI(), "application/json", body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func (cv *CallbackValidator) Middleware(c *gin.Context) error {
	token := &credential.CallbackToken{
		Bucket:   c.GetHeader("Bucket"),
		FileName: c.Param("name"),
		Version:  util.ToInt(c.Query("version")),
		Method:   c.Request.Method,
		Extra:    make(map[string][]string),
	}
	for _, key := range cv.cfg.Params {
		arr, _ := c.GetQueryArray(key)
		if val := c.GetHeader(key); val != "" {
			arr = append(arr, val)
		}
		token.Extra[key] = arr
	}
	return cv.Verify(token)
}
