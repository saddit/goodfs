package repo

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/webapi"
	"common/logs"
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
func (m *MetadataRepo) FindByName(name string) *entity.Metadata {
	//FIXME: load balance with slaves
	servs := logic.NewDiscovery().GetMetaServers(true)
	loc, err := logic.NewHashSlot().FindLocOfName(name, servs)
	if err != nil {
		logs.Std().Errorf("find metadata by name error: %s", err)
		return nil
	}
	metadata, err := webapi.GetMetadata(loc, name)
	if err != nil {
		logs.Std().Errorf("find metadata by name error: %s", err)
		return nil
	}
	return metadata
}

//FindByNameAndVerMode 根据文件名查找元数据 verMode筛选版本数据
func (m *MetadataRepo) FindByNameAndVerMode(name string, verMode entity.VerMode) *entity.Metadata {
	metadata := m.FindByName(name)
	//TODO 根据VerMode同时查询版本
	switch verMode {
	case entity.VerModeALL:
	case entity.VerModeLast:
	case entity.VerModeNot:
	default:
	}
	return metadata
}

func (m *MetadataRepo) Insert(data *entity.Metadata) (*entity.Metadata, error) {
	masters := logic.NewDiscovery().GetMetaServers(true)
	loc, err := logic.NewHashSlot().FindLocOfName(data.Name, masters)
	if err != nil {
		return nil, err
	}
	if err := webapi.PostMetadata(loc, *data); err != nil {
		return nil, err
	}
	return data, nil
}
