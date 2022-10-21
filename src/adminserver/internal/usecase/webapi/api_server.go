package webapi

import (
	"adminserver/internal/usecase/pool"
	"common/response"
	"fmt"
	"io"
	"net/http"
)

func ListVersion(ip, name string, page, pageSize int) ([]byte, error) {
	resp, err := pool.Http.Get(fmt.Sprintf("http://%s/metadata/%s/versions?page=%d&page_size=%d", ip, name, page, pageSize))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return io.ReadAll(resp.Body)
}
