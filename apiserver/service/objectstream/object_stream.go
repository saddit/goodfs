package objectstream

import (
	"fmt"
	"io"
	"net/http"
)

var client = http.Client{}

type PutStream struct {
	Locate    string
	name      string
	writer    *io.PipeWriter
	errorChan chan error
}

type GetStream struct {
	Locate string
	reader io.ReadCloser
}

func NewPutStream(ip, name string, size int64) *PutStream {
	reader, writer := io.Pipe()
	c := make(chan error, 1)
	go func() {
		req, _ := http.NewRequest(http.MethodPut, "http://"+ip+"/objects/"+name, reader)
		req.Header.Add("Size", fmt.Sprint(size))
		resp, e := client.Do(req)
		if resp.StatusCode != http.StatusOK {
			e = fmt.Errorf("dataServer return http code %v", resp.StatusCode)
		}
		c <- e
	}()
	return &PutStream{writer: writer, errorChan: c, Locate: ip, name: name}
}

func NewGetStream(ip, name string) (*GetStream, error) {
	resp, err := client.Get("http://" + ip + "/objects/" + name)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dataServer return http code %v", resp.StatusCode)
	}
	return &GetStream{ip, resp.Body}, nil
}

func (r *GetStream) Read(b []byte) (int, error) {
	return r.reader.Read(b)
}

func (r *GetStream) Close() error {
	return r.reader.Close()
}

func (p *PutStream) Close() error {
	defer close(p.errorChan)
	err := p.writer.Close()
	if err != nil {
		return err
	}
	return <-p.errorChan
}

func (p *PutStream) Write(b []byte) (n int, err error) {
	return p.writer.Write(b)
}

//Commit send commit message and close stream
func (p *PutStream) Commit(ok bool) error {
	if e := p.Close(); e != nil {
		return e
	}
	if !ok {
		go DeleteObject(p.Locate, p.name)
	}
	return nil
}
