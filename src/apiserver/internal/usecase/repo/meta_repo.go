package repo

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/grpcapi"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/webapi"
	"fmt"
)

type MetadataRepo struct{}

func NewMetadataRepo() *MetadataRepo {
	return &MetadataRepo{}
}

// FindByName 根据文件名查找元数据 不查询版本
func (m *MetadataRepo) FindByName(name, bucket string, withExtra bool) (*entity.Metadata, error) {
	name = fmt.Sprint(bucket, "/", name)
	masterId, err := logic.NewHashSlot().KeySlotLocation(name)
	if err != nil {
		return nil, err
	}
	ip, err := logic.NewDiscovery().SelectMetaServerGRPC(masterId)
	if err != nil {
		return nil, err
	}
	return grpcapi.GetMetadata(ip, name, withExtra)
}

func (m *MetadataRepo) Insert(data *entity.Metadata) error {
	name := fmt.Sprint(data.Bucket, "/", data.Name)
	masterId, err := logic.NewHashSlot().KeySlotLocation(name)
	if err != nil {
		return err
	}
	loc := logic.NewDiscovery().GetMetaServerHTTP(masterId)
	return webapi.PostMetadata(loc, *data)
}
