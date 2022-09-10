package repo

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/webapi"
	"common/logs"
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	ErrVersion int32 = -1
)

type VersionRepo struct {
	kv clientv3.KV
}

func NewVersionRepo(kv clientv3.KV) *VersionRepo {
	return &VersionRepo{kv}
}

//Find 根据文件名字和版本号返回版本元数据
func (v *VersionRepo) Find(name string, version int32) *entity.Version {
	//TODO 从etcd获取元数据所在节点的位置
	// 从etcd元数据所在节点位置，选取一个可用节点发送获取版本的请求
	return nil
}

//Update updating locate and setting ts to now
func (v *VersionRepo) Update(ctx context.Context, ver *entity.Version) bool {
	//TODO 从etcd元数据所在节点位置，向主节点发送覆盖版本元数据的请求
	// Locate不为空则覆盖etcd中数据分片位置
	return false
}

//Add 为metadata添加一个版本，添加到版本数组的末尾，版本号为数组序号
//返回对应版本号,如果失败返回ErrVersion -1
func (v *VersionRepo) Add(ctx context.Context, name string, ver *entity.Version) int32 {
	masters := logic.NewDiscovery().GetMetaServers(true)
	loc, err := logic.NewHashSlot().FindLocOfName(name, masters)
	if err != nil {
		return ErrVersion
	}
	verNum, err := webapi.PostVersion(loc, name, ver)
	if err != nil {
		logs.Std().Error(err)
		return ErrVersion
	}
	return int32(verNum)
}

func (v *VersionRepo) Delete(ctx context.Context, name string, ver *entity.Version) error {
	//TODO 从etcd获取元数据所在节点位置，向主节点发送删除版本元数据的请求
	return nil
}
