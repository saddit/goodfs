package webapi

import (
	"adminserver/internal/usecase/pool"
	"common/response"
	"common/util"
	"fmt"
	"io"
	"net/http"
)

func ListVersion(ip, name string, page, pageSize int) ([]byte, int, error) {
	resp, err := pool.Http.Get(fmt.Sprintf("%s://%s/metadata/%s/versions?page=%d&page_size=%d", GetSchema(), ip, name, page, pageSize))
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	total := util.ToInt(resp.Header.Get("X-Total-Count"))
	bt, err := io.ReadAll(resp.Body)
	return bt, total, err
}

func PutObjects(ip, name, sha256 string, fileIO io.Reader) error {
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s://%s/objects/%s", GetSchema(), ip, name), fileIO)
	if err != nil {
		return err
	}
	req.Header.Set("Digest", sha256)
	resp, err := pool.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func GetObjects(ip, name string, version int) (io.ReadCloser, error) {
	resp, err := pool.Http.Get(fmt.Sprintf("%s://%s/objects/%s?version=%d", GetSchema(), ip, name, version))
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
