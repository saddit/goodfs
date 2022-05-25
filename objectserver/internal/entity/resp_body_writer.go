package entity

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

type RespBodyWriter struct {
	gin.ResponseWriter
	Body *bytes.Buffer
}

func (r RespBodyWriter) Write(b []byte) (int, error) {
	r.Body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r RespBodyWriter) Read(p []byte) (int, error) {
	return r.Body.Read(p)
}

func (r RespBodyWriter) Close() error {
	return nil
}
