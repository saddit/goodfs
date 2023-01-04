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
	n, err := r.Writer.Write(b)
	if n > 0 {
		if n, err = r.Body.Write(b[:n]); err != nil {
			return n, err
		}
	}
	return n, err
}

type BufferTeeReader struct {
	io.Reader
	Body *bytes.Buffer
}

func (r *BufferTeeReader) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)
	if n > 0 {
		if n, err = r.Body.Write(b[:n]); err != nil {
			return n, err
		}
	}
	return n, err
}
