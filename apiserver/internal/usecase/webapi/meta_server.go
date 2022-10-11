package webapi

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/pool"
	"bytes"
	"common/request"
	"common/response"
	"common/util"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetMetadata(ip, name string, verNum int32) (*entity.Metadata, error) {
	resp, err := pool.Http.Get(fmt.Sprint(metaRest(ip, name), "?version=", verNum))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return util.UnmarshalPtrFromIO[entity.Metadata](resp.Body)
}

func PostMetadata(ip string, data entity.Metadata) error {
	data.Versions = nil
	bt, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	resp, err := pool.Http.Post(metaRest(ip, ""), request.ContentTypeJSON, bytes.NewBuffer(bt))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func PutMetadata(ip string, data entity.Metadata) error {
	data.Versions = nil
	bt, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	req, err := request.GetPutReq(bytes.NewBuffer(bt), metaRest(ip, data.Name), request.ContentTypeJSON)
	if err != nil {
		return err
	}
	resp, err := pool.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func DelMetadata(ip, name string) error {
	req, err := request.GetDeleteReq(metaRest(ip, name))
	if err != nil {
		return err
	}
	resp, err := pool.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func GetVersion(ip, name string, verNum int32) (*entity.Version, error) {
	resp, err := pool.Http.Get(versionNumRest(ip, name, verNum))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return util.UnmarshalPtrFromIO[entity.Version](resp.Body)
}

func ListVersion(ip, name string, page, pageSize int) ([]*entity.Version, error) {
	resp, err := pool.Http.Get(versionListRest(ip, name, page, pageSize))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return util.UnmarshalFromIO[[]*entity.Version](resp.Body)
}

func PostVersion(ip, name string, body *entity.Version) (uint64, error) {
	bt, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}
	resp, err := pool.Http.Post(metaVerRest(ip, name), request.ContentTypeJSON, bytes.NewBuffer(bt))
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusCreated {
		return 0, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return util.ToUint64(resp.Header.Get("Version")), nil
}

func PutVersion(ip, name string, body *entity.Version) error {
	bt, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := request.GetPutReq(bytes.NewBuffer(bt), versionNumRest(ip, name, body.Sequence), request.ContentTypeJSON)
	if err != nil {
		return err
	}
	resp, err := pool.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

// DelVersion verNum < 0 will delete all version
func DelVersion(ip, name string, verNum int32) error {
	req, err := request.GetDeleteReq(versionNumRest(ip, name, verNum))
	if err != nil {
		return err
	}
	resp, err := pool.Http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func metaRest(ip, name string) string {
	if name == "" {
		return fmt.Sprintf("http://%s/metadata", ip)
	}
	return fmt.Sprintf("http://%s/metadata/%s", ip, name)
}

func metaVerRest(ip, name string) string {
	if name == "" {
		return fmt.Sprintf("http://%s/metadata_version", ip)
	}
	return fmt.Sprintf("http://%s/metadata_version/%s", ip, name)
}

func versionListRest(ip, name string, page, pageSize int) string {
	return fmt.Sprintf("http://%s/metadata_version/%s?page=%d&page_size=%d", ip, name, page, pageSize)
}

func versionNumRest(ip, name string, num int32) string {
	return fmt.Sprintf("http://%s/metadata_version/%s?version=%d", ip, name, num)
}
