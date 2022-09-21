package service

import (
	"common/util"
	"errors"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
)

type MetadataService struct {
	repo IMetadataRepo
	batch IBatchMetaRepo
	cache IMetaCache
}

func NewMetadataService(repo IMetadataRepo, batch IBatchMetaRepo, c IMetaCache) *MetadataService {
	return &MetadataService{repo, batch, c}
}

func (m *MetadataService) AddMetadata(data *entity.Metadata) error {
	if ok, resp :=  m.repo.ApplyRaft(&entity.RaftData{
		Type: entity.LogInsert,
		Dest: entity.DestMetadata,
		Name: data.Name,
		Metadata: data,
	}); ok {
		return resp
	}

	return m.repo.AddMetadata(data)
}

func (m *MetadataService) AddVersion(name string, data *entity.Version) (int, error) {
	if ok, resp := m.repo.ApplyRaft(&entity.RaftData{
		Type: entity.LogInsert,
		Dest: entity.DestVersion,
		Name: name,
		Version: data,
	}); ok {
		if resp.Ok() { return int(resp.Data.(uint64)), nil }
		return -1, nil
	}

	if err := m.repo.AddVersion(name, data); err != nil {
		return -1, err
	}
	return int(data.Sequence), nil
}

func (m *MetadataService) UpdateMetadata(name string, data *entity.Metadata) error {
	if ok, resp := m.repo.ApplyRaft(&entity.RaftData{
		Type: entity.LogUpdate,
		Dest: entity.DestMetadata,
		Name: name,
		Metadata: data,
	}); ok {
		return resp
	}

	return m.repo.UpdateMetadata(name, data)
}

func (m *MetadataService) UpdateVersion(name string, ver int, data *entity.Version) error {
	data.Sequence = uint64(ver)
	if ok, resp := m.repo.ApplyRaft(&entity.RaftData{
		Type: entity.LogUpdate,
		Dest: entity.DestVersion,
		Name: name,
		Sequence: data.Sequence,
		Version: data,
	}); ok {
		return resp
	}

	return m.repo.UpdateVersion(name, data)
}

func (m *MetadataService) RemoveMetadata(name string) error {
	if ok, resp := m.repo.ApplyRaft(&entity.RaftData{
		Type: entity.LogRemove,
		Dest: entity.DestMetadata,
		Name: name,
	}); ok {
		return resp
	}
	
	return m.repo.RemoveMetadata(name);
}

func (m *MetadataService) RemoveVersion(name string, ver int) error {
	if ok, resp := m.repo.ApplyRaft(&entity.RaftData{
		Type: entity.LogRemove,
		Dest: util.IfElse(ver < 0, entity.DestVersionAll, entity.DestVersion),
		Name: name,
		Sequence: uint64(ver),
	}); ok {
		return resp
	}

	if ver < 0 {
		return m.repo.RemoveAllVersion(name)
	} else {
		return m.repo.RemoveVersion(name, uint64(ver))
	}
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
		if errors.Is(err, ErrNotFound) {
			return meta, nil, nil
		} else if err != nil {
			return nil, nil, err
		}
		return meta, ver, nil
	}
}

func (m *MetadataService) GetVersion(name string, ver int) (*entity.Version, error) {
	if ver <= 0 {
		return m.repo.GetVersion(name, m.repo.GetLastVersionNumber(name))
	}
	return m.repo.GetVersion(name, uint64(ver))
}

func (m *MetadataService) ListVersions(name string, page int, size int) ([]*entity.Version, error) {
	if page == 0 {
		page = 1
	}
	start := (page - 1) * size
	return m.repo.ListVersions(name, start, start+size)
}

// FilterKeys heavy!
func (m *MetadataService) FilterKeys(fn func(string) bool) []string {
	var keys []string
	m.batch.ForeachKeys(func (key string) bool {
		if fn(key) {
			keys = append(keys, key)
		}
		return true
	})
	return keys
}

func (m *MetadataService) ForeachVersionBytes(name string, fn func([]byte) bool) {
	m.repo.ForeachVersionBytes(name, fn)
}

func (m *MetadataService) GetMetadataBytes(name string) ([]byte, error) {
	return m.repo.GetMetadataBytes(name)
}