package service

import (
	"apiserver/internal/usecase/webapi"
	"common/logs"
	"fmt"
	"io"
	"net/http"
)

type GetStream struct {
	reader io.ReadCloser
	Locate string
	name   string
	offset int
	size   int64
}

// NewGetStream IO: Head object
func NewGetStream(ip, name string, size int64) (*GetStream, error) {
	stream := &GetStream{
		reader: nil,
		Locate: ip,
		name:   name,
		size:   size,
	}
	return stream, stream.CheckStat()
}

func (g *GetStream) CheckStat() error {
	return webapi.HeadObject(g.Locate, g.name)
}

func (g *GetStream) request() error {
	resp, err := webapi.GetObject(g.Locate, g.name, g.offset, g.size)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get object from dataServer return http code %v", resp.StatusCode)
	}
	g.reader = resp.Body
	return nil
}

func (g *GetStream) Seek(offset int64, whence int) (int64, error) {
	if whence != io.SeekStart {
		logs.Std().Warn("get stream only supports seek whence io.SeekStart")
	}
	g.offset = int(offset)
	return offset, nil
}

func (g *GetStream) Read(bt []byte) (int, error) {
	if g.reader == nil {
		if err := g.request(); err != nil {
			return 0, err
		}
	}
	return g.reader.Read(bt)
}

func (g *GetStream) Close() error {
	if g.reader == nil {
		return nil
	}
	return g.reader.Close()
}
