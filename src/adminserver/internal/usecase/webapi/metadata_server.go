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

func ListMetadata(ip, prefix string, pageSize int, orderBy string, desc bool) ([]*entity.Metadata, error) {
	resp, err := pool.Http.Get(metadataListRest(ip, map[string][]string{
		"prefix":    {prefix},
		"page_size": {util.ToString(pageSize)},
		"order_by":  {util.ToString(orderBy)},
		"desc":      {util.ToString(desc)},
	}))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return util.UnmarshalFromIO[[]*entity.Metadata](resp.Body)
}

func metadataListRest(ip string, param map[string][]string) string {
	return fmt.Sprintf("http://%s/metadata/list?%s", ip, url.Values(param).Encode())
}
