package webapi

import (
	"apiserver/internal/entity"
	"common/request"
	"common/response"
	"common/util"
	"fmt"
	"net/http"
)

func GetBucket(ip, name string) (*entity.Bucket, error) {
	defer perform(false)()
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/bucket/%s", ip, name), nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return util.UnmarshalPtrFromIO[entity.Bucket](resp.Body)
}

func PutBucket(ip string, data *entity.Bucket) error {
	defer perform(true)()
	req, err := request.JsonReq(http.MethodPut, fmt.Sprintf("http://%s/bucket/%s", ip, data.Name), data)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func PostBucket(ip string, data *entity.Bucket) error {
	defer perform(true)()
	req, err := request.JsonReq(http.MethodPost, fmt.Sprintf("http://%s/bucket", ip), data)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func DeleteBucket(ip, name string) error {
	defer perform(true)()
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s/bucket/%s", ip, name), nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func ListBucket(ip, name, prefix string, size int) error {
	defer perform(false)()
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/bucket/%s?page_size=%d&prefix=%s", ip, name, prefix, size), nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}
