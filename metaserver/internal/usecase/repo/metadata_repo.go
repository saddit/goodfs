package repo

import (
	"bytes"
	"common/graceful"
	"common/logs"
	"common/response"
	"common/util"
	"fmt"
	"io"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/logic"
	"metaserver/internal/usecase/pool"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"
)

type MetadataRepo struct {
	*db.Storage
}

func NewMetadataRepo(db *db.Storage) *MetadataRepo {
	return &MetadataRepo{Storage: db}
}

func (m *MetadataRepo) ApplyRaft(data *entity.RaftData) (bool, *response.RaftFsmResp) {
	if rf, ok := pool.RaftWrapper.GetRaftIfLeader(); ok {
		bt, err := util.EncodeMsgp(data)
		if err != nil {
			return true, response.NewRaftFsmResp(err)
		}
		feat := rf.Apply(bt, 5*time.Second)
		if err := feat.Error(); err != nil {
			return true, response.NewRaftFsmResp(err)
		}
		if resp := feat.Response(); resp != nil {
			return true, resp.(*response.RaftFsmResp)
		}
		return true, nil
	}
	return false, nil
}

func (m *MetadataRepo) AddMetadata(data *entity.Metadata) error {
	if data == nil {
		return ErrNilData
	}
	return m.Update(logic.AddMeta(data))
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

func (m *MetadataRepo) RemoveVersion(name string, ver uint64) error {
	return m.Update(logic.RemoveVer(name, ver))
}

func (m *MetadataRepo) RemoveAllVersion(name string) error {
	return m.Update(func(tx *bolt.Tx) error {
		root := logic.GetRoot(tx)
		// delete bucket
		if err := root.DeleteBucket([]byte(name)); err != nil {
			return err
		}
		// create an emtpy bucket
		_, err := root.CreateBucket([]byte(name))
		return err
	})
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
	return data, m.DB().View(logic.GetVer(name, ver, data))
}

func (m *MetadataRepo) ListVersions(name string, start int, end int) (lst []*entity.Version, err error) {
	lst = make([]*entity.Version, 0, end-start+1)
	err = m.DB().View(func(tx *bolt.Tx) error {
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
			if err := util.DecodeMsgp(data, v); err != nil {
				return err
			}
			lst = append(lst, data)
		}

		return nil
	})
	return
}

func (m *MetadataRepo) ReadDB() (io.ReadCloser, error) {
	reader, writer := io.Pipe()
	errCh := make(chan error)
	go func() {
		defer graceful.Recover()
		defer writer.Close()
		tx, err := m.DB().Begin(false)
		if err != nil {
			errCh <- err
			close(errCh)
			return
		}
		defer tx.Rollback()
		close(errCh)
		n, err := tx.WriteTo(writer)
		if err != nil {
			logs.Std().Error("writer (ReadDB) error: %v, written %d", err, n)
			return
		}
	}()
	if err := <-errCh; err != nil {
		return nil, err
	}
	return reader, nil
}

func (m *MetadataRepo) ReplaceDB(r io.Reader) (err error) {
	dbPath := m.DB().Path() + "_replace"
	// open new db file
	newFile, err := os.OpenFile(dbPath, os.O_WRONLY|os.O_CREATE, util.OS_ModeUser)
	if err != nil {
		logs.Std().Error("restore fail on open new file: %v", err)
		return err
	}
	// save new db data
	n, err := io.Copy(newFile, r)
	if err != nil {
		logs.Std().Error("restore fail on copy data to new file: %v, written %d", err, n)
		return err
	}
	if err := newFile.Close(); err != nil {
		logs.Std().Error("close new db file err: %s", err)
		return err
	}
	// reopen db
	return m.Replace(dbPath)
}
