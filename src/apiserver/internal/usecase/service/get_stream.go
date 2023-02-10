package service

import (
	"apiserver/internal/usecase/webapi"
	"fmt"
	"io"
	"net/http"
)

type GetStream struct {
	reader   io.ReadCloser
	Locate   string
	name     string
	size     int64
	compress bool
}

// NewGetStream IO: Head object
func NewGetStream(ip, name string, size int64, compress bool) (*GetStream, error) {
	stream := &GetStream{
		reader:   nil,
		Locate:   ip,
		name:     name,
		size:     size,
		compress: compress,
	}
	return stream, stream.CheckStat()
}

func (g *GetStream) CheckStat() error {
	return webapi.HeadObject(g.Locate, g.name)
}

func (g *GetStream) request(offset int) error {
	resp, err := webapi.GetObject(g.Locate, g.name, offset, g.size, g.compress)
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
	if offset < 0 {
		return 0, fmt.Errorf("get stream only supports forward seek offest")
	}
	if whence == io.SeekEnd {
		return 0, fmt.Errorf("get stream only supports SeekStart and SeekCurrent")
	}
	if g.reader == nil {
		if err := g.request(int(offset)); err != nil {
			return 0, err
		}
		return offset, nil
	}
	if offset > 0 {
		n, err := io.ReadFull(g.reader, make([]byte, offset))
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err = nil
		}
		return int64(n), err
	}
	return 0, nil
}

func (g *GetStream) Read(bt []byte) (int, error) {
	if g.reader == nil {
		if err := g.request(0); err != nil {
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
