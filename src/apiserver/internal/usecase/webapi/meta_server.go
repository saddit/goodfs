package webapi

import (
	"apiserver/internal/entity"
	"bytes"
	"common/pb"
	"common/request"
	"common/response"
	"common/util"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func GetMetadata(ip, name string, verNum int32, withExtra bool) (*entity.Metadata, error) {
	defer perform(false)()
	resp, err := httpClient.Get(fmt.Sprintf("%s?version=%d&with_extra=%t", metaRest(ip, name), verNum, withExtra))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return util.UnmarshalPtrFromIO[entity.Metadata](resp.Body)
}

func PostMetadata(ip string, data entity.Metadata) error {
	defer perform(true)()
	data.Versions = nil
	bt, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	resp, err := httpClient.Post(metaRest(ip, url.PathEscape(fmt.Sprint(data.Bucket, "/", data.Name))), request.ContentTypeJSON, bytes.NewBuffer(bt))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func PutMetadata(ip string, data entity.Metadata) error {
	defer perform(true)()
	data.Versions = nil
	bt, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	id := url.PathEscape(fmt.Sprint(data.Bucket, "/", data.Name))
	req, err := request.GetPutReq(bytes.NewBuffer(bt), metaRest(ip, id), request.ContentTypeJSON)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func DelMetadata(ip, name string) error {
	defer perform(true)()
	req, err := request.GetDeleteReq(metaRest(ip, name))
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func ListMetadata(ip, prefix string, pageSize int) ([]*entity.Metadata, error) {
	defer perform(false)()
	resp, err := httpClient.Get(metadataListRest(ip, prefix, pageSize))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return util.UnmarshalFromIO[[]*entity.Metadata](resp.Body)
}

func GetVersion(ip, name string, verNum int32) (*entity.Version, error) {
	defer perform(false)()
	resp, err := httpClient.Get(versionNumRest(ip, name, verNum))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return util.UnmarshalPtrFromIO[entity.Version](resp.Body)
}

func ListVersion(ip, id string, page, pageSize int) ([]byte, int, error) {
	defer perform(false)()
	resp, err := httpClient.Get(versionListRest(ip, id, page, pageSize))
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	defer resp.Body.Close()
	total := util.ToInt(resp.Header.Get("X-Total-Count"))
	bt, err := io.ReadAll(resp.Body)
	return bt, total, err
}

func PostVersion(ip, id string, body *entity.Version) (uint64, error) {
	defer perform(true)()
	bt, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}
	resp, err := httpClient.Post(versionRest(ip, id), request.ContentTypeJSON, bytes.NewBuffer(bt))
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusCreated {
		return 0, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return util.ToUint64(resp.Header.Get("Version")), nil
}

func PutVersion(ip, id string, body *entity.Version) error {
	defer perform(true)()
	bt, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := request.GetPutReq(bytes.NewBuffer(bt), versionNumRest(ip, id, body.Sequence), request.ContentTypeJSON)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
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
	defer perform(true)()
	req, err := request.GetDeleteReq(versionNumRest(ip, name, verNum))
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}

func VersionsByHash(ip, hash string) ([]*pb.Version, error) {
	defer perform(false)()
	resp, err := httpClient.Get(fmt.Sprintf("http://%s/version/list?hash=%s", ip, hash))
	if err != nil {
		return nil, err
	}
	bt, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res []*pb.Version
	if err = json.Unmarshal(bt, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func metaRest(ip, name string) string {
	if name == "" {
		return fmt.Sprintf("http://%s/metadata", ip)
	}
	return fmt.Sprintf("http://%s/metadata/%s", ip, name)
}

func versionRest(ip, name string) string {
	if name == "" {
		return fmt.Sprintf("http://%s/metadata_version", ip)
	}
	return fmt.Sprintf("http://%s/metadata_version/%s", ip, name)
}

func metadataListRest(ip, prefix string, pageSize int) string {
	return fmt.Sprintf("http://%s/metadata/list?page_size=%d&prefix=%s", ip, pageSize, prefix)
}

func versionListRest(ip, name string, page, pageSize int) string {
	return fmt.Sprintf("http://%s/metadata_version/%s/list?page=%d&page_size=%d", ip, name, page, pageSize)
}

func versionNumRest(ip, name string, num int32) string {
	return fmt.Sprintf("http://%s/metadata_version/%s?version=%d", ip, name, num)
}
