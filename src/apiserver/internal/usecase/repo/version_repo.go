package repo

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/grpcapi"
	"apiserver/internal/usecase/logic"
	"fmt"
)

const (
	ErrVersion int32 = -1
)

type VersionRepo struct {
}

func NewVersionRepo() *VersionRepo {
	return &VersionRepo{}
}

// Find return the metadata of specified version
func (v *VersionRepo) Find(name, bucket string, version int32) (*entity.Version, error) {
	name = fmt.Sprint(bucket, "/", name)
	masterId, err := logic.NewHashSlot().KeySlotLocation(name)
	if err != nil {
		return nil, err
	}
	ip, err := logic.NewDiscovery().SelectMetaServerGRPC(masterId)
	if err != nil {
		return nil, err
	}
	return grpcapi.GetVersion(ip, name, version)
}

// Update updating locate and setting ts to now
func (v *VersionRepo) Update(name, bucket string, ver *entity.Version) error {
	name = fmt.Sprint(bucket, "/", name)
	masterId, err := logic.NewHashSlot().KeySlotLocation(name)
	if err != nil {
		return err
	}
	return grpcapi.UpdateVersion(logic.NewDiscovery().GetMetaServerGRPC(masterId), name, ver)
}

// Add add a version for metadata. returns the num of version
func (v *VersionRepo) Add(name, bucket string, ver *entity.Version) (int32, error) {
	name = fmt.Sprint(bucket, "/", name)
	masterId, err := logic.NewHashSlot().KeySlotLocation(name)
	if err != nil {
		return ErrVersion, err
	}
	ver.Sequence, err = grpcapi.SaveVersion(logic.NewDiscovery().GetMetaServerGRPC(masterId), name, ver)
	if err != nil {
		return ErrVersion, err
	}
	return ver.Sequence, nil
}

func (v *VersionRepo) Delete(name, bucket string, ver int32) error {
	name = fmt.Sprint(bucket, "/", name)
	masterId, err := logic.NewHashSlot().KeySlotLocation(name)
	if err != nil {
		return err
	}
	return grpcapi.RemoveVersion(logic.NewDiscovery().GetMetaServerHTTP(masterId), name, ver)
}
