package service

import (
	"metaserver/internal/usecase/logic"
	"common/proto/msg"
	"common/util"
	"errors"
	"metaserver/internal/entity"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/raftimpl"
	"strings"
	"time"
)

type MetadataService struct {
	usecase.RaftApply
	repo      usecase.IMetadataRepo
	batch     usecase.IBatchMetaRepo
	hashIndex usecase.IHashIndexRepo
}

func NewMetadataService(repo usecase.IMetadataRepo, batch usecase.IBatchMetaRepo, hashIndex usecase.IHashIndexRepo, rw *raftimpl.RaftWrapper) *MetadataService {
	return &MetadataService{raftimpl.RaftApplier(rw), repo, batch, hashIndex}
}

func (m *MetadataService) AddMetadata(id string, data *msg.Metadata) error {
	data.CreateTime = time.Now().UnixMilli()
	data.UpdateTime = data.CreateTime
	if ok, _, err := m.ApplyRaft(&entity.RaftData{
		Type:     entity.LogInsert,
		Dest:     entity.DestMetadata,
		Name:     id,
		Metadata: data,
	}); ok {
		return err
	}

	return m.repo.AddMetadata(id, data)
}

func (m *MetadataService) AddVersion(name string, data *msg.Version) (int, error) {
	data.UniqueId = logic.GenerateUniqueId()
	data.Ts = time.Now().UnixMilli()
	if ok, resp, err := m.ApplyRaft(&entity.RaftData{
		Type:    entity.LogInsert,
		Dest:    entity.DestVersion,
		Name:    name,
		Version: data,
	}); ok {
		if err != nil {
			return -1, err
		}
		return int(resp.(uint64)), nil
	}

	if err := m.repo.AddVersion(name, data); err != nil {
		return -1, err
	}
	return int(data.Sequence), nil
}

func (m *MetadataService) ReceiveVersion(name string, data *msg.Version) error {
	if ok, _, err := m.ApplyRaft(&entity.RaftData{
		Type:    entity.LogMigrate,
		Dest:    entity.DestVersion,
		Name:    name,
		Version: data,
	}); ok {
		return err
	}

	if err := m.repo.AddVersionFromRaft(name, data); err != nil {
		return err
	}
	return nil
}

func (m *MetadataService) UpdateMetadata(name string, data *msg.Metadata) error {
	data.UpdateTime = time.Now().UnixMilli()
	if ok, _, err := m.ApplyRaft(&entity.RaftData{
		Type:     entity.LogUpdate,
		Dest:     entity.DestMetadata,
		Name:     name,
		Metadata: data,
	}); ok {
		return err
	}

	return m.repo.UpdateMetadata(name, data)
}

func (m *MetadataService) UpdateVersion(name string, ver int, data *msg.Version) error {
	data.Ts = time.Now().UnixMilli()
	data.Sequence = uint64(ver)
	if ok, _, err := m.ApplyRaft(&entity.RaftData{
		Type:     entity.LogUpdate,
		Dest:     entity.DestVersion,
		Name:     name,
		Sequence: data.Sequence,
		Version:  data,
	}); ok {
		return err
	}

	return m.repo.UpdateVersion(name, data)
}

func (m *MetadataService) RemoveMetadata(name string) error {
	if ok, _, err := m.ApplyRaft(&entity.RaftData{
		Type: entity.LogRemove,
		Dest: entity.DestMetadata,
		Name: name,
	}); ok {
		return err
	}

	return m.repo.RemoveMetadata(name)
}

func (m *MetadataService) RemoveVersion(name string, ver int) error {
	if ok, _, err := m.ApplyRaft(&entity.RaftData{
		Type:     entity.LogRemove,
		Dest:     util.IfElse(ver < 0, entity.DestVersionAll, entity.DestVersion),
		Name:     name,
		Sequence: uint64(ver),
	}); ok {
		return err
	}

	if ver < 0 {
		return m.repo.RemoveAllVersion(name)
	} else {
		return m.repo.RemoveVersion(name, uint64(ver))
	}
}

// GetMetadata 获取metadata及其版本，如果version为-1则不获取任何版本，返回的版本为nil
func (m *MetadataService) GetMetadata(id string, version int, withExtra bool) (*msg.Metadata, *msg.Version, error) {
	meta, err := m.repo.GetMetadata(id)
	if err != nil {
		return nil, nil, err
	}
	if withExtra {
		extra, err := m.repo.GetExtra(id)
		if err != nil {
			return nil, nil, err
		}
		meta.Extra = extra
	}
	switch version {
	case -1:
		return meta, nil, nil
	default:
		ver, err := m.GetVersion(id, version)
		if err != nil && !errors.Is(err, usecase.ErrNotFound) {
			return nil, nil, err
		}
		return meta, ver, nil
	}
}

func (m *MetadataService) GetVersion(name string, ver int) (*msg.Version, error) {
	if ver <= 0 {
		return m.repo.GetVersion(name, m.repo.GetLastVersionNumber(name))
	}
	return m.repo.GetVersion(name, uint64(ver))
}

func (m *MetadataService) ListVersions(name string, page int, size int) ([]*msg.Version, int, error) {
	if page == 0 {
		page = 1
	}
	// start at 1
	start := (page-1)*size + 1
	return m.repo.ListVersions(name, start, start+size)
}

func (m *MetadataService) ListMetadata(prefix string, size int) ([]*msg.Metadata, int, error) {
	if size == 0 {
		return []*msg.Metadata{}, 0, nil
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

func (m *MetadataService) FindByHash(hash string) (res []*msg.Version, err error) {
	keys, err := m.hashIndex.FindAll(hash)
	if err != nil {
		return nil, err
	}
	res = make([]*msg.Version, 0, len(keys))
	needSync := false
	for _, key := range keys {
		idx := strings.LastIndexByte(key, '.')
		name, sequence := key[0:idx], util.ToInt(key[idx+1:])
		ver, err := m.GetVersion(name, sequence)
		if errors.Is(err, usecase.ErrNotFound) {
			needSync = true
			_ = m.hashIndex.Remove(hash, key)
			continue
		} else if err != nil {
			return nil, err
		}
		res = append(res, ver)
	}
	if needSync {
		err = m.hashIndex.Sync()
	}
	return
}

func (m *MetadataService) UpdateLocates(hash string, index int, locate string) error {
	return m.repo.UpdateLocateByHash(hash, index, locate)
}
