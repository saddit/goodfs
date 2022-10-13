package auth

import (
	"bytes"
	"common/response"
	"net/http"
	"net/url"
)

type CallbackValidator struct {
	cfg *CallbackConfig
	cli *http.Client
}

func NewCallbackValidator(cli *http.Client, cfg *Config) *CallbackValidator {
	return &CallbackValidator{cfg: &cfg.Callback, cli: cli}
}

func (cv *CallbackValidator) Verify(token Credential) error {
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
