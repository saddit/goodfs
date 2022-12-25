package service

import (
	"apiserver/internal/usecase/webapi"
	"fmt"
	"io"
	"net/http"
)

type GetStream struct {
	reader io.ReadCloser
	Locate string
	name   string
}

//NewGetStream IO: Head object
func NewGetStream(ip, name string) (*GetStream, error) {
	stream := &GetStream{nil, ip, name}
	return stream, stream.CheckStat()
}

func (g *GetStream) CheckStat() error {
	return webapi.HeadObject(g.Locate, g.name)
}

func (g *GetStream) request() error {
	resp, err := webapi.GetObject(g.Locate, g.name)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get object from dataServer return http code %v", resp.StatusCode)
	}
	g.reader = resp.Body
	return nil
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
	return g.reader.Close()
}
