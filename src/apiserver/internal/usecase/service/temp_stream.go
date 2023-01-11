package service

import (
	"apiserver/internal/usecase/webapi"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type tempStream struct {
	reader io.ReadCloser
	Locate string
	name   string
	size   int64
}

// NewTempStream IO: Head object
func NewTempStream(ip, name string, size int64) *tempStream {
	stream := &tempStream{
		Locate: ip,
		name:   name,
		size:   size,
	}
	return stream
}

func (ts *tempStream) CheckStat() error {
	_, err := webapi.HeadTmpObject(ts.Locate, ts.name)
	return err
}

func (ts *tempStream) request() error {
	resp, err := webapi.GetTmpObject(ts.Locate, ts.name, ts.size)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get temp object from dataServer return http code %v", resp.StatusCode)
	}
	ts.reader = resp.Body
	return nil
}

func (ts *tempStream) Seek(int64, int) (int64, error) {
	return 0, errors.New("temp stream not support seek")
}

func (ts *tempStream) Read(bt []byte) (int, error) {
	if ts.reader == nil {
		if err := ts.request(); err != nil {
			return 0, err
		}
	}
	return ts.reader.Read(bt)
}

func (ts *tempStream) Close() error {
	if ts.reader == nil {
		return nil
	}
	return ts.reader.Close()
}
