package repo

import (
	"apiserver/internal/entity"

	clientv3 "go.etcd.io/etcd/client/v3"
	// log "github.com/sirupsen/logrus"
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
	var data entity.Metadata
	//TODO 从etcd获取Metadata所在节点的位置 选择一个可用节点发送获取元数据Header的请求
	return &data
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
