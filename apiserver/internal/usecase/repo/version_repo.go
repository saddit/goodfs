package repo

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/webapi"
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
func (v *VersionRepo) Find(name string, version int32) (*entity.Version, error) {
	//FIXME: load balance with slaves
	servs := logic.NewDiscovery().GetMetaServers(true)
	loc, err := logic.NewHashSlot().FindLocOfName(name, servs)
	if err != nil {
		return nil, err
	}
	return webapi.GetVersion(loc, name, version)
}

//Update updating locate and setting ts to now
func (v *VersionRepo) Update(name string, ver *entity.Version) error {
	masters := logic.NewDiscovery().GetMetaServers(true)
	loc, err := logic.NewHashSlot().FindLocOfName(name, masters)
	if err != nil {
		return err
	}
	return webapi.PutVersion(loc, name, ver)
}

//Add 为metadata添加一个版本
//返回对应版本号,如果失败返回ErrVersion -1
func (v *VersionRepo) Add(name string, ver *entity.Version) (int32, error) {
	masters := logic.NewDiscovery().GetMetaServers(true)
	loc, err := logic.NewHashSlot().FindLocOfName(name, masters)
	if err != nil {
		return ErrVersion, err
	}
	verNum, err := webapi.PostVersion(loc, name, ver)
	if err != nil {
		return ErrVersion, err
	}
	ver.Sequence = int32(verNum)
	return ver.Sequence, nil
}

func (v *VersionRepo) Delete(name string, ver *entity.Version) error {
	masters := logic.NewDiscovery().GetMetaServers(true)
	loc, err := logic.NewHashSlot().FindLocOfName(name, masters)
	if err != nil {
		return err
	}
	return webapi.DelVersion(loc, name, ver.Sequence)
}
