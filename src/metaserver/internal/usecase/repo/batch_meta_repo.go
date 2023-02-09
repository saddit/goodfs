package repo

import (
	"common/proto/msg"
	bolt "go.etcd.io/bbolt"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/logic"
)

type BatchMetaRepo struct {
	Storage *db.Storage
}

func NewBatchRepo(stroe *db.Storage) *BatchMetaRepo {
	return &BatchMetaRepo{Storage: stroe}
}

func (br *BatchMetaRepo) Sync() error {
	return br.Storage.DB().Sync()
}

func (br *BatchMetaRepo) ForeachKeys(fn func(string) bool) {
	_ = br.Storage.View(logic.ForeachKeys(fn))
}

func (br *BatchMetaRepo) AddVersion(name string, data *msg.Version) error {
	if data == nil {
		return usecase.ErrNilData
	}
	return br.Storage.DB().Batch(logic.AddVer(name, data))
}

func (br *BatchMetaRepo) UpdateVersion(name string, data *msg.Version) error {
	if data == nil {
		return usecase.ErrNilData
	}
	return br.Storage.DB().Batch(logic.UpdateVer(name, data))
}

func (br *BatchMetaRepo) RemoveVersion(name string, ver uint64) error {
	return br.Storage.DB().Batch(logic.RemoveVer(name, ver))
}

func (br *BatchMetaRepo) AddMetadata(id string, data *msg.Metadata) error {
	if data == nil {
		return usecase.ErrNilData
	}
	return br.Storage.DB().Batch(logic.AddMeta(id, data))
}

func (br *BatchMetaRepo) UpdateMetadata(name string, data *msg.Metadata) error {
	return br.Storage.DB().Batch(logic.UpdateMeta(name, data))
}

func (br *BatchMetaRepo) RemoveMetadata(name string) error {
	return br.Storage.DB().Batch(logic.RemoveMeta(name))
}

func (br *BatchMetaRepo) AddVersionWithSequence(id string, data *msg.Version) error {
	if data == nil {
		return usecase.ErrNilData
	}
	return br.Storage.DB().Batch(logic.AddVerWithSequence(id, data))
}

func (br *BatchMetaRepo) RemoveAllVersion(id string) error {
	return br.Storage.DB().Batch(func(tx *bolt.Tx) error {
		// delete bucket
		if err := logic.RemoveVersionBucket(tx, id); err != nil {
			return err
		}
		// create an empty bucket
		return logic.CreateVersionBucket(tx, id)
	})
}
