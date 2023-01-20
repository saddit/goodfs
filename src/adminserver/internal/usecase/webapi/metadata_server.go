package webapi

import (
	"adminserver/internal/entity"
	"adminserver/internal/usecase/pool"
	"common/response"
	"common/util"
	"fmt"
	"net/http"
	"net/url"
)

func ListMetadata(ip, prefix string, pageSize int) ([]*entity.Metadata, int, error) {
	resp, err := pool.Http.Get(metadataListRest(ip, map[string][]string{
		"prefix":    {prefix},
		"page_size": {util.ToString(pageSize)},
	}))
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	total := util.ToInt(resp.Header.Get("X-Total-Count"))
	lst, err := util.UnmarshalFromIO[[]*entity.Metadata](resp.Body)
	return lst, total, err
}

func ListBuckets(ip, prefix string, pageSize int) ([]*entity.Bucket, int, error) {
	resp, err := pool.Http.Get(bucketListRest(ip, map[string][]string{
		"prefix":    {prefix},
		"page_size": {util.ToString(pageSize)},
	}))
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	total := util.ToInt(resp.Header.Get("X-Total-Count"))
	lst, err := util.UnmarshalFromIO[[]*entity.Bucket](resp.Body)
	return lst, total, err
}

func metadataListRest(ip string, param map[string][]string) string {
	return fmt.Sprintf("http://%s/metadata/list?%s", ip, url.Values(param).Encode())
}

func bucketListRest(ip string, param map[string][]string) string {
	return fmt.Sprintf("http://%s/bucket/list?%s", ip, url.Values(param).Encode())
}
