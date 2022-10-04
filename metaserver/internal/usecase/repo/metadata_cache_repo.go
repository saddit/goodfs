package repo

import (
	"common/cache"
	"common/util"
	"errors"
	"fmt"
	"metaserver/internal/entity"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/logic"
)

type MetadataCacheRepo struct {
	cache cache.ICache
}

func NewMetadataCacheRepo(c cache.ICache) *MetadataCacheRepo {
	return &MetadataCacheRepo{c}
}

func (m *MetadataCacheRepo) GetMetadata(s string) (*entity.Metadata, error) {
	if bt, ok := m.cache.HasGet(s); ok {
		var en entity.Metadata
		if err := util.DecodeMsgp(&en, bt); err != nil {
			return nil, err
		}
		return &en, nil
	}
	return nil, usecase.ErrNotFound
}

func (m *MetadataCacheRepo) GetVersion(s string, u uint64) (*entity.Version, error) {
	key := fmt.Sprint(s, logic.Sep, u)
	if bt, ok := m.cache.HasGet(key); ok {
		var en entity.Version
		if err := util.DecodeMsgp(&en, bt); err != nil {
			return nil, err
		}
		return &en, nil
	}
	return nil, usecase.ErrNotFound
}

// ListVersions return successfully matched cache until failure.
// if error is not nil, error string is the started version should be fetched from db
func (m *MetadataCacheRepo) ListVersions(s string, start int, end int) ([]*entity.Version, error) {
	size := end - start + 1
	res := make([]*entity.Version, 0, size)
	for i := start; i <= end; i++ {
		v, err := m.GetVersion(s, uint64(i))
		if err != nil {
			return res, errors.New(fmt.Sprint(i))
		}
		res = append(res, v)
	}
	return res, nil
}

func (m *MetadataCacheRepo) AddMetadata(metadata *entity.Metadata) error {
	bt, err := util.EncodeMsgp(metadata)
	if err != nil {
		return err
	}
	m.cache.Set(metadata.Name, bt)
	return nil
}

func (m *MetadataCacheRepo) AddVersion(s string, version *entity.Version) error {
	key := fmt.Sprint(s, logic.Sep, version.Sequence)
	bt, err := util.EncodeMsgp(version)
	if err != nil {
		return err
	}
	m.cache.Set(key, bt)
	return nil
}

func (m *MetadataCacheRepo) UpdateMetadata(s string, metadata *entity.Metadata) error {
	metadata.Name = s
	return m.AddMetadata(metadata)
}

func (m *MetadataCacheRepo) UpdateVersion(s string, version *entity.Version) error {
	return m.AddVersion(s, version)
}

func (m *MetadataCacheRepo) RemoveMetadata(s string) error {
	m.cache.Delete(s)
	return nil
}

func (m *MetadataCacheRepo) RemoveVersion(s string, u uint64) error {
	m.cache.Delete(fmt.Sprint(s, logic.Sep, u))
	return nil
}
