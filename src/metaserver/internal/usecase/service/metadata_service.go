package service

import (
	"common/pb"
	"common/util"
	"errors"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
	"strings"
)

type MetadataService struct {
	repo      IMetadataRepo
	batch     IBatchMetaRepo
	hashIndex IHashIndexRepo
}

func NewMetadataService(repo IMetadataRepo, batch IBatchMetaRepo, hashIndex IHashIndexRepo) *MetadataService {
	return &MetadataService{repo, batch, hashIndex}
}

func (m *MetadataService) AddMetadata(data *entity.Metadata) error {
	if ok, resp := m.repo.ApplyRaft(&entity.RaftData{
		Type:     entity.LogInsert,
		Dest:     entity.DestMetadata,
		Name:     data.Name,
		Metadata: data,
	}); ok {
		return resp
	}

	return m.repo.AddMetadata(data)
}

func (m *MetadataService) AddVersion(name string, data *entity.Version) (int, error) {
	if ok, resp := m.repo.ApplyRaft(&entity.RaftData{
		Type:    entity.LogInsert,
		Dest:    entity.DestVersion,
		Name:    name,
		Version: data,
	}); ok {
		if resp.Ok() {
			return int(resp.Data.(uint64)), nil
		}
		return -1, nil
	}

	if err := m.repo.AddVersion(name, data); err != nil {
		return -1, err
	}
	return int(data.Sequence), nil
}

func (m *MetadataService) ReceiveVersion(name string, data *entity.Version) error {
	if ok, resp := m.repo.ApplyRaft(&entity.RaftData{
		Type:    entity.LogMigrate,
		Dest:    entity.DestVersion,
		Name:    name,
		Version: data,
	}); ok {
		if resp.Ok() {
			return nil
		}
		return nil
	}

	if err := m.repo.AddVersionWithSequnce(name, data); err != nil {
		return err
	}
	return nil
}

func (m *MetadataService) UpdateMetadata(name string, data *entity.Metadata) error {
	if ok, resp := m.repo.ApplyRaft(&entity.RaftData{
		Type:     entity.LogUpdate,
		Dest:     entity.DestMetadata,
		Name:     name,
		Metadata: data,
	}); ok {
		return resp
	}

	return m.repo.UpdateMetadata(name, data)
}

func (m *MetadataService) UpdateVersion(name string, ver int, data *entity.Version) error {
	data.Sequence = uint64(ver)
	if ok, resp := m.repo.ApplyRaft(&entity.RaftData{
		Type:     entity.LogUpdate,
		Dest:     entity.DestVersion,
		Name:     name,
		Sequence: data.Sequence,
		Version:  data,
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

	return m.repo.RemoveMetadata(name)
}

func (m *MetadataService) RemoveVersion(name string, ver int) error {
	if ok, resp := m.repo.ApplyRaft(&entity.RaftData{
		Type:     entity.LogRemove,
		Dest:     util.IfElse(ver < 0, entity.DestVersionAll, entity.DestVersion),
		Name:     name,
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
		if err != nil && !errors.Is(err, ErrNotFound) {
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
	// start at 1
	start := (page-1)*size + 1
	return m.repo.ListVersions(name, start, start+size)
}

func (m *MetadataService) ListMetadata(prefix string, size int) ([]*entity.Metadata, error) {
	if size == 0 {
		return []*entity.Metadata{}, nil
	}
	return m.repo.ListMetadata(prefix, size)
}

// FilterKeys heavy!
func (m *MetadataService) FilterKeys(fn func(string) bool) []string {
	var keys []string
	m.batch.ForeachKeys(func(key string) bool {
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

func (m *MetadataService) FindByHash(hash string) (res []*pb.Version, err error) {
	keys, err := m.hashIndex.FindAll(hash)
	if err != nil {
		return nil, err
	}
	res = make([]*pb.Version, 0, len(keys))
	needSync := false
	for _, key := range keys {
		idx := strings.LastIndexByte(key, '.')
		name, sequence := key[0:idx], util.ToInt(key[idx+1:])
		ver, err := m.GetVersion(name, sequence)
		if errors.Is(err, ErrNotFound) {
			needSync = true
			_ = m.hashIndex.Remove(hash, key)
			continue
		} else if err != nil {
			return nil, err
		}
		res = append(res, &pb.Version{
			Hash:      ver.Hash,
			Sequence:  ver.Sequence,
			Size:      ver.Size,
			Name:      name,
			Locations: ver.Locate,
		})
	}
	if needSync {
		err = m.hashIndex.Sync()
	}
	return
}

func (m *MetadataService) UpdateLocates(name string, version int, locates []string) error {
	ver, err := m.GetVersion(name, version)
	if err != nil {
		return err
	}
	ver.Locate = locates
	return m.UpdateVersion(name, version, ver)
}
