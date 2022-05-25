package service

import (
	"apiserver/internal/usecase/webapi"
	"fmt"
	"io"
	"net/http"
)

type GetStream struct {
	io.ReadCloser
	Locate string
}

//NewGetStream IO: Get object
func NewGetStream(ip, name string) (*GetStream, error) {
	resp, err := webapi.GetObject(ip, name)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dataServer return http code %v", resp.StatusCode)
	}
	return &GetStream{resp.Body, ip}, nil
}
