package repo

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/webapi"
	"fmt"
	"net/url"
)

type MetadataRepo struct {
	versionRepo IVersionRepo
}

func NewMetadataRepo(vr IVersionRepo) *MetadataRepo {
	return &MetadataRepo{vr}
}

//FindByName 根据文件名查找元数据 不查询版本
func (m *MetadataRepo) FindByName(name, bucket string) (*entity.Metadata, error) {
	name = fmt.Sprint(bucket, "/", name)
	loc, gid, err := logic.NewHashSlot().FindMetaLocByName(name)
	if err != nil {
		return nil, err
	}
	loc = logic.NewDiscovery().SelectMetaByGroupID(gid, loc)
	return webapi.GetMetadata(loc, url.PathEscape(name), int32(entity.VerModeNot), false)
}

//FindByNameWithVersion 根据文件名查找元数据 verMode筛选版本数据
func (m *MetadataRepo) FindByNameWithVersion(name, bucket string, verMode entity.VerMode, withExtra bool) (*entity.Metadata, error) {
	name = fmt.Sprint(bucket, "/", name)
	loc, gid, err := logic.NewHashSlot().FindMetaLocByName(name)
	if err != nil {
		return nil, err
	}
	loc = logic.NewDiscovery().SelectMetaByGroupID(gid, loc)
	return webapi.GetMetadata(loc, url.PathEscape(name), int32(verMode), withExtra)
}

func (m *MetadataRepo) Insert(data *entity.Metadata) (*entity.Metadata, error) {
	name := fmt.Sprint(data.Bucket, "/", data.Name)
	loc, _, err := logic.NewHashSlot().FindMetaLocByName(name)
	if err != nil {
		return nil, err
	}
	if err = webapi.PostMetadata(loc, *data); err != nil {
		return nil, err
	}
	if len(data.Versions) > 0 {
		data.Versions[0].Sequence, err = m.versionRepo.Add(data.Name, data.Bucket, data.Versions[0])
	}
	return data, err
}
