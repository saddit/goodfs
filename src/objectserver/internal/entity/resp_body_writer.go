package entity

import (
	"bytes"
	"io"
)

type BufferTeeWriter struct {
	io.Writer
	Body *bytes.Buffer
}

func (r *BufferTeeWriter) Write(b []byte) (int, error) {
	r.Body.Write(b)
	return r.Writer.Write(b)
}

type BufferTeeReader struct {
	io.Reader
	Body *bytes.Buffer
}

func (r *BufferTeeReader) Read(b []byte) (int, error) {
	r.Body.Write(b)
	return r.Reader.Read(b)
}
