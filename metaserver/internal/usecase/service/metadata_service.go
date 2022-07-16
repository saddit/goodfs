package service

import (
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
)

type MetadataService struct {
	repo IMetadataRepo
}

func NewMetadataService(repo IMetadataRepo) *MetadataService {
	return &MetadataService{repo}
}

func (m *MetadataService) AddMetadata(name string, data *entity.Metadata) error {
	r := m.repo
	return r.AddMetadata(name, data)
}

func (m *MetadataService) AddVersion(name string, data *entity.Version) (int, error) {
	err := m.repo.AddVersion(name, data)
	if err != nil {
		return -1, err
	}
	return int(data.Sequence), nil
}

func (m *MetadataService) UpdateMetadata(name string, data *entity.Metadata) error {
	return m.repo.UpdateMetadata(name, data)
}

func (m *MetadataService) UpdateVersion(name string, ver int, data *entity.Version) error {
	data.Sequence = uint64(ver)
	return m.repo.UpdateVersion(name, data)
}

func (m *MetadataService) RemoveMetadata(name string) error {
	return m.repo.RemoveMetadata(name)
}

func (m *MetadataService) RemoveVersion(name string, ver int) error {
	return m.repo.RemoveVersion(name, ver)
}

// GetMetadata 获取metadata及其版本，如果version为-1则不获取任何版本，返回的版本为nil
func (m *MetadataService) GetMetadata(name string, version int) (*entity.Metadata, *entity.Version, error) {
	meta, err := m.repo.GetMetadata(name)
	if err != nil {
		return nil, nil, err
	}
	switch version {
	case -1:
		return meta, nil, nil
	default:
		ver, err := m.GetVersion(name, version)
		if err != nil {
			return nil, nil, err
		}
		return meta, ver, nil
	}
}

func (m *MetadataService) GetVersion(name string, ver int) (*entity.Version, error) {
	return m.repo.GetVersion(name, uint64(ver))
}

func (m *MetadataService) ListVersions(name string, start int, end int) ([]*entity.Version, error) {
	return m.repo.ListVersions(name, start, end)
}
