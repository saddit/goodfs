package repo

import (
	"bytes"
	"errors"
	"fmt"
	"metaserver/internal/entity"
	"time"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"github.com/tinylib/msgp/msgp"
)

const (
	BucketName = "metadata"
	Sep        = "."
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("already exists key")
	ErrNilData  = errors.New("nil data")
	ErrDecode   = errors.New("decode fail")
	ErrEncode   = errors.New("encode fail")
)

type MetadataRepo struct {
	*bolt.DB
}

func NewMetadataRepo(db *bolt.DB) *MetadataRepo {
	return &MetadataRepo{db}
}

func (m *MetadataRepo) ExistMetadata(name string) (exist bool) {
	_ = m.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		if b == nil {
			return nil
		}
		if b.Get([]byte(name)) != nil {
			exist = true
		}
		return nil
	})
	return 
}

func (m *MetadataRepo) AddMetadata(name string, data *entity.Metadata) error {
	if data == nil {
		return ErrNilData
	}
	return m.Update(func(tx *bolt.Tx) error {
		bucket, _ := tx.CreateBucketIfNotExists([]byte(BucketName))
		key := []byte(name)
		if bucket.Get(key) != nil {
			return ErrExists
		}
		bt := encodeMsg(data)
		if bt == nil {
			return ErrDecode
		}
		data.CreateTime = time.Now().Unix()
		data.UpdateTime = time.Now().Unix()
		if err := bucket.Put(key, bt); err != nil {
			return err
		}

		return nil
	})
}

func (m *MetadataRepo) UpdateMetadata(name string, data *entity.Metadata) error {
	return nil
}

func (m *MetadataRepo) RemoveMetadata(name string) error {
	return m.Update(func(tx *bolt.Tx) error {
		key := []byte(name)
		if err := tx.DeleteBucket(key); err != nil {
			return err
		}
		if bucket := tx.Bucket([]byte(BucketName)); bucket != nil {
			return bucket.Delete(key)
		}
		return ErrNotFound
	})
}

func (m *MetadataRepo) GetMetadata(name string) (*entity.Metadata, error) {
	data := &entity.Metadata{}
	return data, m.View(func(tx *bolt.Tx) error {
		if bucket := tx.Bucket([]byte(BucketName)); bucket != nil {
			bt := bucket.Get([]byte(name))
			if bt == nil {
				return ErrNotFound
			}
			if !decodeMsg(data, bt) {
				return ErrDecode
			}
			return nil
		}
		return ErrNotFound
	})
}

func (m *MetadataRepo) AddVersion(name string, data *entity.Version) error {
	if data == nil {
		return ErrNilData
	}
	return m.Update(func(tx *bolt.Tx) error {
		if bucket := tx.Bucket([]byte(name)); bucket != nil {
			data.Sequence, _ = bucket.NextSequence()
			data.Ts = time.Now().Unix()
			key := []byte(fmt.Sprint(name, Sep, data.Sequence))
			if bt := encodeMsg(data); bt != nil {
				return bucket.Put(key, bt)
			}
			return ErrEncode
		}
		return ErrNotFound
	})
}

func (m *MetadataRepo) UpdateVersion(name string, data *entity.Version) error {
	if data == nil {
		return ErrNilData
	}
	return m.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(name)); b != nil {
			key := []byte(fmt.Sprint(name, Sep, data.Sequence))
			data.Ts = time.Now().Unix()
			bt := encodeMsg(data)
			if bt == nil {
				return ErrEncode
			}
			return b.Put(key, bt)
		}
		return ErrNotFound
	})
}

func (m *MetadataRepo) RemoveVersion(name string, ver int) error {
	return m.Update(func(tx *bolt.Tx) error {
		key := []byte(fmt.Sprint(name, Sep, ver))
		b := tx.Bucket([]byte(name))
		if b != nil {
			return ErrNotFound
		}
		if err := b.Delete(key); err != nil {
			return ErrNotFound
		}
		return nil
	})
}

func (m *MetadataRepo) GetVersion(name string, ver int) (*entity.Version, error) {
	data := &entity.Version{}
	return data, m.View(func(tx *bolt.Tx) error {
		if bucket := tx.Bucket([]byte(name)); bucket != nil {
			bt := bucket.Get([]byte(fmt.Sprint(name, Sep, ver)))
			if bt == nil {
				return ErrNotFound
			}
			if !decodeMsg(data, bt) {
				return ErrDecode
			}
			return nil
		}
		return ErrNotFound
	})
}

func (m *MetadataRepo) ListVersions(name string, start int, end int) (lst []*entity.Version, err error) {
	lst = make([]*entity.Version, 0, end-start+1)
	err = m.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(name)).Cursor()

		min := []byte(fmt.Sprint(name, Sep, start))
		max := []byte(fmt.Sprint(name, Sep, end))

		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			data := &entity.Version{}
			if !decodeMsg(data, v) {
				return ErrDecode
			}
			lst = append(lst, data)
		}

		return nil
	})
	return
}


// decodeMsg decode data by msgp if error return false
func decodeMsg[T msgp.Unmarshaler](data T, bt []byte) bool {
	if _, err := data.UnmarshalMsg(bt); err != nil {
		logrus.Errorf("%T decode err: %v", data, err)
		return false
	}
	return true
}

// encodeMsg encode data with msgp if error return nil
func encodeMsg(data msgp.MarshalSizer) []byte {
	bt, err := data.MarshalMsg(nil)
	if err != nil {
		logrus.Errorf("%T encode err: %v", data, err)
		return nil
	}
	return bt
}