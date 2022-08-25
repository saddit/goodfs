package repo

import (
	"bytes"
	"common/logs"
	"fmt"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/logic"
	"metaserver/internal/usecase/pool"
	"metaserver/internal/usecase/utils"
	"time"

	bolt "go.etcd.io/bbolt"
)

type MetadataRepo struct {
	*db.Storage
}

func NewMetadataRepo(db *db.Storage) *MetadataRepo {
	return &MetadataRepo{Storage: db}
}

func (m *MetadataRepo) ApplyRaft(data *entity.RaftData) error {
	if rf, ok := pool.RaftWrapper.GetRaftIfLeader(); ok {
		bt := utils.EncodeMsgp(data)
		if bt == nil {
			return ErrEncode
		}
		feat := rf.Apply(bt, 5*time.Second)
		if err := feat.Error(); err != nil {
			return err
		}
		if resp := feat.Response(); resp != nil {
			return resp.(error)
		}
	}
	return nil
}

func (m *MetadataRepo) AddMetadata(name string, data *entity.Metadata) error {
	if data == nil {
		return ErrNilData
	}
	return m.Update(logic.AddMeta(name, data))
}

func (m *MetadataRepo) UpdateMetadata(name string, data *entity.Metadata) error {
	return m.Update(logic.UpdateMeta(name, data))
}

func (m *MetadataRepo) RemoveMetadata(name string) error {
	return m.Update(logic.RemoveMeta(name))
}

func (m *MetadataRepo) GetMetadata(name string) (*entity.Metadata, error) {
	data := &entity.Metadata{}
	return data, m.View(logic.GetMeta(name, data))
}

func (m *MetadataRepo) AddVersion(name string, data *entity.Version) error {
	if data == nil {
		return ErrNilData
	}
	return m.Update(logic.AddVer(name, data))
}

func (m *MetadataRepo) UpdateVersion(name string, data *entity.Version) error {
	if data == nil {
		return ErrNilData
	}
	return m.Update(logic.UpdateVer(name, data))
}

func (m *MetadataRepo) RemoveVersion(name string, ver int) error {
	return m.Update(logic.RemoveVer(name, ver))
}

func (m *MetadataRepo) GetLastVersionNumber(name string) uint64 {
	var max uint64 = 1
	if err := m.View(func(tx *bolt.Tx) error {
		max = logic.GetRootNest(tx, name).Sequence()
		return nil
	}); err != nil {
		logs.Std().Errorf("GetLastVersionNumber: %+v", err)
	}
	return max
}

func (m *MetadataRepo) GetVersion(name string, ver uint64) (*entity.Version, error) {
	data := &entity.Version{}
	return data, m.View(logic.GetVer(name, ver, data))
}

func (m *MetadataRepo) ListVersions(name string, start int, end int) (lst []*entity.Version, err error) {
	lst = make([]*entity.Version, 0, end-start+1)
	err = m.View(func(tx *bolt.Tx) error {
		root, _ := tx.CreateBucketIfNotExists([]byte(logic.BucketName))
		buk := root.Bucket([]byte(name))
		if buk == nil {
			return nil
		}
		c := buk.Cursor()

		min := []byte(fmt.Sprint(name, logic.Sep, start))
		max := []byte(fmt.Sprint(name, logic.Sep, end))

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			data := &entity.Version{}
			if !utils.DecodeMsgp(data, v) {
				return ErrDecode
			}
			lst = append(lst, data)
		}

		return nil
	})
	return
}
