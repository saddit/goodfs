package repo

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/webapi"
	"fmt"
	"net/url"
)

const (
	ErrVersion int32 = -1
)

type VersionRepo struct {
}

func NewVersionRepo() *VersionRepo {
	return &VersionRepo{}
}

//Find return the metadata of specified version
func (v *VersionRepo) Find(name, bucket string, version int32) (*entity.Version, error) {
	name = fmt.Sprint(bucket, "/", name)
	loc, gid, err := logic.NewHashSlot().FindMetaLocByName(name)
	if err != nil {
		return nil, err
	}
	loc = logic.NewDiscovery().SelectMetaByGroupID(gid, loc)
	return webapi.GetVersion(loc, url.PathEscape(name), version)
}

//Update updating locate and setting ts to now
func (v *VersionRepo) Update(name, bucket string, ver *entity.Version) error {
	name = fmt.Sprint(bucket, "/", name)
	loc, _, err := logic.NewHashSlot().FindMetaLocByName(name)
	if err != nil {
		return err
	}
	return webapi.PutVersion(loc, url.PathEscape(name), ver)
}

// Add add a version for metadata. returns the num of version
func (v *VersionRepo) Add(name, bucket string, ver *entity.Version) (int32, error) {
	name = fmt.Sprint(bucket, "/", name)
	loc, _, err := logic.NewHashSlot().FindMetaLocByName(name)
	if err != nil {
		return ErrVersion, err
	}
	verNum, err := webapi.PostVersion(loc, url.PathEscape(name), ver)
	if err != nil {
		return ErrVersion, err
	}
	ver.Sequence = int32(verNum)
	return ver.Sequence, nil
}

func (v *VersionRepo) Delete(name, bucket string, ver int32) error {
	name = fmt.Sprint(bucket, "/", name)
	loc, _, err := logic.NewHashSlot().FindMetaLocByName(name)
	if err != nil {
		return err
	}
	return webapi.DelVersion(loc, url.PathEscape(name), ver)
}
