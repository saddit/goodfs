package webapi

import (
	"adminserver/internal/entity"
	"adminserver/internal/usecase/pool"
	"common/request"
	"common/response"
	"common/util"
	"fmt"
	"io"
	"net/http"
)

func ListVersion(ip, name, bucket string, page, pageSize int, token string) ([]byte, int, error) {
	url := fmt.Sprintf("%s://%s/v1/metadata/%s/versions?page=%d&page_size=%d", GetSchema(), ip, name, page, pageSize)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Bucket", bucket)
	req.Header.Set("Authorization", token)
	resp, err := pool.Http.Do(req)
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

func PutObjects(ip, name, sha256, bucket string, fileIO io.Reader, size int64, token string) error {
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s://%s/v1/objects/%s", GetSchema(), ip, name), fileIO)
	if err != nil {
		return err
	}
	req.ContentLength = size
	req.Header.Set("Bucket", bucket)
	req.Header.Set("Authorization", token)
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

func GetObjects(ip, name, bucket string, version int, token string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s/v1/objects/%s?version=%d", GetSchema(), ip, name, version), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Bucket", bucket)
	req.Header.Set("Authorization", token)
	resp, err := pool.Http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return resp.Body, nil
}

func CheckToken(ip, token string) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s/v1/security/check", GetSchema(), ip), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	resp, err := pool.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func CreateBucket(ip string, b *entity.Bucket, token string) error {
	req, err := request.JsonReq(http.MethodPost, fmt.Sprintf("%s://%s/v1/bucket", GetSchema(), ip), b)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	resp, err := pool.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func UpdateBucket(ip string, b *entity.Bucket, token string) error {
	req, err := request.JsonReq(http.MethodPut, fmt.Sprintf("%s://%s/v1/bucket/%s", GetSchema(), ip, b.Name), b)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	resp, err := pool.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func DeleteBucket(ip, name string, token string) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s/v1/bucket/%s", GetSchema(), ip, name), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	resp, err := pool.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}
