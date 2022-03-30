package objectstream

import (
	"fmt"
	"io"
	"net/http"
)

var client = http.Client{}

type PutStream struct {
	writer    *io.PipeWriter
	errorChan chan error
}

type GetStream struct {
	reader io.ReadCloser
}

func NewPutStream(ip, name string) *PutStream {
	reader, writer := io.Pipe()
	c := make(chan error)
	go func() {
		req, _ := http.NewRequest("Put", "http://"+ip+"/objects/"+name, reader)
		resp, e := client.Do(req)
		if e == nil && resp.StatusCode != http.StatusOK {
			e = fmt.Errorf("dataServer return http code %v", resp.StatusCode)
		}
		c <- e
	}()
	return &PutStream{writer, c}
}

func NewGetStream(ip, name string) (*GetStream, error) {
	resp, err := client.Get("http://" + ip + "/objects/" + name)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dataServer return http code %v", resp.StatusCode)
	}
	return &GetStream{resp.Body}, nil
}

func (r *GetStream) Read(b []byte) (int, error) {
	return r.reader.Read(b)
}

func (r *GetStream) Close() error {
	return r.reader.Close()
}

func (p *PutStream) Close() error {
	p.writer.Close()
	return <-p.errorChan
}

func (p *PutStream) Write(b []byte) (n int, err error) {
	return p.writer.Write(b)
}
