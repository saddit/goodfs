package repo

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/selector"
	"apiserver/internal/usecase/webapi"
	"common/logs"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type MetadataRepo struct {
	kv clientv3.KV
	versionRepo IVersionRepo
}

func NewMetadataRepo(kv clientv3.KV, vr IVersionRepo) *MetadataRepo {
	return &MetadataRepo{kv, vr}
}

//FindByName 根据文件名查找元数据 不查询版本
func (m *MetadataRepo) FindByName(name string) *entity.Metadata {
	//TODO 从etcd获取所在集群
	servs := logic.NewDiscovery().GetMetaServers(false)
	lb := selector.NewIPSelector(pool.Balancer, servs)
	
	metadata, err := webapi.GetMetadata(lb.Select(), name)
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
	//TODO 选取一个主元数据节点，保存元数据到这主节点，记录位置信息到etcd
	// 如果Version不为空则保存Version元数据
	return data, nil
}
