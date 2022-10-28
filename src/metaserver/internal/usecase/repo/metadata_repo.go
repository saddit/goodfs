package repo

import (
	"common/constrant"
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
	MainDB *db.Storage
	Cache  IMetaCache
}

func NewMetadataRepo(db *db.Storage, c IMetaCache) *MetadataRepo {
	return &MetadataRepo{MainDB: db, Cache: c}
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
	if err := m.MainDB.Update(logic.AddMeta(data)); err != nil {
		return err
	}
	go func() {
		defer graceful.Recover()
		err := m.Cache.AddMetadata(data)
		util.LogErrWithPre("metadata cache", err)
	}()
	return nil
}

func (m *MetadataRepo) UpdateMetadata(name string, data *entity.Metadata) error {
	if err := m.MainDB.Update(logic.UpdateMeta(name, data)); err != nil {
		return err
	}
	go func() {
		defer graceful.Recover()
		err := m.Cache.UpdateMetadata(name, data)
		util.LogErrWithPre("metadata cache", err)
	}()
	return nil
}

func (m *MetadataRepo) RemoveMetadata(name string) error {
	if err := m.MainDB.Update(logic.RemoveMeta(name)); err != nil {
		return err
	}
	go func() {
		defer graceful.Recover()
		err := m.Cache.RemoveMetadata(name)
		util.LogErrWithPre("metadata cache", err)
	}()
	return nil
}

func (m *MetadataRepo) GetMetadata(name string) (*entity.Metadata, error) {
	if data, err := m.Cache.GetMetadata(name); err == nil {
		return data, nil
	}
	data := &entity.Metadata{}
	if err := m.MainDB.View(logic.GetMeta(name, data)); err != nil {
		return nil, err
	}
	go func() {
		defer graceful.Recover()
		util.LogErrWithPre("add metadata cache", m.Cache.AddMetadata(data))
	}()
	return data, nil
}

func (m *MetadataRepo) AddVersion(name string, data *entity.Version) error {
	if data == nil {
		return ErrNilData
	}
	if err := m.MainDB.Update(logic.AddVer(name, data)); err != nil {
		return err
	}
	go func() {
		defer graceful.Recover()
		err := m.Cache.AddVersion(name, data)
		util.LogErrWithPre("metadata cache", err)
	}()
	return nil
}

func (m *MetadataRepo) UpdateVersion(name string, data *entity.Version) error {
	if data == nil {
		return ErrNilData
	}
	if err := m.MainDB.Update(logic.UpdateVer(name, data)); err != nil {
		return err
	}
	go func() {
		defer graceful.Recover()
		err := m.Cache.UpdateVersion(name, data)
		util.LogErrWithPre("metadata cache", err)
	}()
	return nil
}

func (m *MetadataRepo) RemoveVersion(name string, ver uint64) error {
	if err := m.MainDB.Update(logic.RemoveVer(name, ver)); err != nil {
		return err
	}
	go func() {
		defer graceful.Recover()
		err := m.Cache.RemoveVersion(name, ver)
		util.LogErrWithPre("metadata cache", err)
	}()
	return nil
}

func (m *MetadataRepo) RemoveAllVersion(name string) error {
	buk := fmt.Sprint(logic.NestPrefix, name)
	last := m.GetLastVersionNumber(name)
	if err := m.MainDB.Update(func(tx *bolt.Tx) error {
		root := logic.GetRoot(tx)
		// delete bucket
		if err := root.DeleteBucket([]byte(buk)); err != nil {
			return err
		}
		// create an empty bucket
		_, err := root.CreateBucket([]byte(buk))
		return err
	}); err != nil {
		return err
	}
	go func() {
		defer graceful.Recover()
		for i := uint64(0); i <= last; i++ {
			err := m.Cache.RemoveVersion(name, i)
			util.LogErrWithPre("remove version cache", err)
		}
	}()
	return nil
}

func (m *MetadataRepo) GetLastVersionNumber(name string) uint64 {
	var max uint64 = 1
	if err := m.MainDB.View(func(tx *bolt.Tx) error {
		if buk := logic.GetRootNest(tx, name); buk != nil {
			max = buk.Sequence()
		}
		return ErrNotFound
	}); err != nil {
		logs.Std().Errorf("GetLastVersionNumber: %+v", err)
	}
	return max
}

func (m *MetadataRepo) GetVersion(name string, ver uint64) (*entity.Version, error) {
	if data, err := m.Cache.GetVersion(name, ver); err == nil {
		return data, nil
	}
	data := &entity.Version{}
	if err := m.MainDB.DB().View(logic.GetVer(name, ver, data)); err != nil {
		return nil, err
	}
	go func() {
		defer graceful.Recover()
		util.LogErrWithPre("add metadata cache", m.Cache.AddVersion(name, data))
	}()
	return data, nil
}

func (m *MetadataRepo) ListVersions(name string, start int, end int) (lst []*entity.Version, err error) {
	size := end - start + 1
	lst, err = m.Cache.ListVersions(name, start, end)
	if err == nil {
		return lst, nil
	}
	start = util.ToInt(err.Error())
	err = m.MainDB.View(func(tx *bolt.Tx) error {
		buk := logic.GetRootNest(tx, name)
		if buk == nil {
			return ErrNotFound
		}
		c := buk.Cursor()

		min := []byte(fmt.Sprint(name, logic.Sep, start))

		for k, v := c.Seek(min); k != nil && len(lst) < size; k, v = c.Next() {
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

func (m *MetadataRepo) ListMetadata(prefix string, size int) (lst []*entity.Metadata, err error) {
	err = m.MainDB.View(func(tx *bolt.Tx) error {
		root := logic.GetRoot(tx)
		if root == nil {
			return ErrNotFound
		}
		cur := root.Cursor()
		var k, v []byte
		if prefix != "" {
			k, v = cur.Seek([]byte(prefix))
		} else {
			k, v = cur.Next()
		}
		for k != nil && len(lst) < size {
			var data entity.Metadata
			if err := util.DecodeMsgp(&data, v); err != nil {
				return err
			}
			lst = append(lst, &data)
			k, v = cur.Next()
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
		tx, err := m.MainDB.DB().Begin(false)
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
	dbPath := m.MainDB.DB().Path() + "_replace"
	// open new db file
	newFile, err := os.OpenFile(dbPath, os.O_WRONLY|os.O_CREATE, constrant.OS.ModeUser)
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
	return m.MainDB.Replace(dbPath)
}

func (m *MetadataRepo) ForeachVersionBytes(name string, fn func([]byte) bool) {
	_ = m.MainDB.View(func(tx *bolt.Tx) error {
		_ = logic.GetRootNest(tx, name).ForEach(func(k, v []byte) error {
			if !fn(v) {
				return ErrNotFound
			}
			return nil
		})
		return nil
	})
}

func (m *MetadataRepo) GetMetadataBytes(key string) ([]byte, error) {
	var res []byte
	err := m.MainDB.View(func(tx *bolt.Tx) error {
		res = logic.GetRoot(tx).Get([]byte(key))
		return nil
	})
	return res, err
}
