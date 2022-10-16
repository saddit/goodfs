package repo

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/webapi"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type MetadataRepo struct {
	kv          clientv3.KV
	versionRepo IVersionRepo
}

func NewMetadataRepo(kv clientv3.KV, vr IVersionRepo) *MetadataRepo {
	return &MetadataRepo{kv, vr}
}

//FindByName 根据文件名查找元数据 不查询版本
func (m *MetadataRepo) FindByName(name string) (*entity.Metadata, error) {
	loc, gid, err := logic.NewHashSlot().FindMetaLocOfName(name)
	if err != nil {
		return nil, err
	}
	loc = logic.NewDiscovery().SelectMetaByGroupID(gid, loc)
	return webapi.GetMetadata(loc, name, int32(entity.VerModeNot))
}

//FindByNameWithVersion 根据文件名查找元数据 verMode筛选版本数据
func (m *MetadataRepo) FindByNameWithVersion(name string, verMode entity.VerMode) (*entity.Metadata, error) {
	loc, gid, err := logic.NewHashSlot().FindMetaLocOfName(name)
	if err != nil {
		return nil, err
	}
	loc = logic.NewDiscovery().SelectMetaByGroupID(gid, loc)
	return webapi.GetMetadata(loc, name, int32(verMode))
}

func (m *MetadataRepo) Insert(data *entity.Metadata) (*entity.Metadata, error) {
	loc, _, err := logic.NewHashSlot().FindMetaLocOfName(data.Name)
	if err != nil {
		return nil, err
	}
	if err = webapi.PostMetadata(loc, *data); err != nil {
		return nil, err
	}
	if len(data.Versions) > 0 {
		data.Versions[0].Sequence, err = m.versionRepo.Add(data.Name, data.Versions[0])
	}
	return data, err
}
