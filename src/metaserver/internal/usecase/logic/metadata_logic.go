package logic

import (
	"common/logs"
	"errors"
	"fmt"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
	"time"

	"common/util"

	bolt "go.etcd.io/bbolt"
)

const (
	MetadataBucketRoot = "go.dfs.metadata.root"
	VersionBucketRoot  = "go.dfs.version.root"
	Sep                = "."
)

func ForeachKeys(fn func(string) bool) TxFunc {
	return func(tx *bolt.Tx) error {
		root := GetMetadataBucket(tx)
		return root.ForEach(func(k, v []byte) error {
			if !fn(string(k)) {
				return ErrNotFound
			}
			return nil
		})
	}
}

func AddMeta(data *entity.Metadata) TxFunc {
	return func(tx *bolt.Tx) error {
		root := GetMetadataBucket(tx)
		key := util.StrToBytes(data.Name)
		// check duplicate
		if root.Get(key) != nil {
			return ErrExists
		}
		data.CreateTime = time.Now().UnixMilli()
		data.UpdateTime = data.CreateTime
		// encode data
		bt, err := util.EncodeMsgp(data)
		if err != nil {
			return err
		}
		// create version bucket
		if err = CreateVersionBucket(tx, data.Name); err != nil {
			return err
		}
		// put metadata
		return root.Put(key, bt)
	}
}

func RemoveMeta(name string) TxFunc {
	return func(tx *bolt.Tx) error {
		key := util.StrToBytes(name)
		root := GetMetadataBucket(tx)
		if root.Get(key) == nil {
			return ErrNotFound
		}
		if err := root.Delete(key); err != nil {
			return err
		}
		err := RemoveVersionBucket(tx, name)
		// ignore err of bucket not found
		if err != nil && !errors.Is(err, bolt.ErrBucketNotFound) {
			return err
		}
		return nil
	}
}

func UpdateMeta(name string, data *entity.Metadata) TxFunc {
	return func(tx *bolt.Tx) error {
		root := GetMetadataBucket(tx)
		var origin entity.Metadata
		if err := getMeta(root, name, &origin); err != nil {
			return err
		}
		if data.UpdateTime < origin.UpdateTime {
			return ErrOldData
		}
		// update data
		data.UpdateTime = time.Now().UnixMilli()
		data.Bucket = origin.Bucket
		bt, err := util.EncodeMsgp(data)
		if err != nil {
			return err
		}
		return root.Put(util.StrToBytes(name), bt)
	}
}

func GetMeta(name string, data *entity.Metadata) TxFunc {
	return func(tx *bolt.Tx) error {
		return getMeta(GetMetadataBucket(tx), name, data)
	}
}

func AddVerWithSequence(name string, data *entity.Version) TxFunc {
	return func(tx *bolt.Tx) error {
		if bucket := GetVersionBucket(tx, name); bucket != nil {
			// only if data is migrated from others will do sequence updating
			if data.Sequence > bucket.Sequence() {
				if err := bucket.SetSequence(data.Sequence); err != nil {
					return fmt.Errorf("set sequence err: %w", err)
				}
			}
			key := util.StrToBytes(fmt.Sprint(name, Sep, data.Sequence))
			if bucket.Get(key) != nil {
				return ErrExists
			}
			data.Ts = time.Now().UnixMilli()
			bt, err := util.EncodeMsgp(data)
			if err != nil {
				return err
			}
			if err := bucket.Put(key, bt); err != nil {
				return err
			}
			return NewHashIndexLogic().AddIndex(data.Hash, string(key))(tx)
		}
		return ErrNotFound
	}
}

func AddVer(name string, data *entity.Version) TxFunc {
	return func(tx *bolt.Tx) error {
		if bucket := GetVersionBucket(tx, name); bucket != nil {
			data.Sequence, _ = bucket.NextSequence()
			key := util.StrToBytes(fmt.Sprint(name, Sep, data.Sequence))
			if bucket.Get(key) != nil {
				return ErrExists
			}
			data.Ts = time.Now().UnixMilli()
			bt, err := util.EncodeMsgp(data)
			if err != nil {
				return err
			}
			if err := bucket.Put(key, bt); err != nil {
				return err
			}
			return NewHashIndexLogic().AddIndex(data.Hash, string(key))(tx)
		}
		return ErrNotFound
	}
}

func RemoveVer(name string, ver uint64) TxFunc {
	return func(tx *bolt.Tx) error {
		key := util.StrToBytes(fmt.Sprint(name, Sep, ver))
		b := GetVersionBucket(tx, name)
		if b == nil {
			return ErrNotFound
		}
		var data entity.Version
		if err := getVer(b, name, ver, &data); err != nil {
			return err
		}
		// remove index
		if err := NewHashIndexLogic().RemoveIndex(data.Hash, string(key))(tx); err != nil {
			return fmt.Errorf("remove hash-index err: %w", err)
		}
		return b.Delete(key)
	}
}

func UpdateVer(name string, data *entity.Version) TxFunc {
	return func(tx *bolt.Tx) error {
		if b := GetVersionBucket(tx, name); b != nil {
			key := util.StrToBytes(fmt.Sprint(name, Sep, data.Sequence))
			// get old one
			var origin entity.Version
			if err := getVer(b, name, data.Sequence, &origin); err != nil {
				return err
			}
			// validate timestamp
			if data.Ts < origin.Ts {
				return ErrOldData
			}
			// those updating are not allowed
			data.Sequence = origin.Sequence
			data.Hash = origin.Hash
			data.Ts = time.Now().UnixMilli()
			// encode to bytes
			bt, err := util.EncodeMsgp(data)
			if err != nil {
				return err
			}
			return b.Put(key, bt)
		}
		return ErrNotFound
	}
}

func GetVer(name string, ver uint64, dest *entity.Version) TxFunc {
	return func(tx *bolt.Tx) error {
		if bucket := GetVersionBucket(tx, name); bucket != nil {
			return getVer(bucket, name, ver, dest)
		}
		return ErrNotFound
	}
}

// GetMetadataBucket get or create metadata root bucket
func GetMetadataBucket(tx *bolt.Tx) *bolt.Bucket {
	if tx.Writable() {
		root, err := tx.CreateBucketIfNotExists(util.StrToBytes(MetadataBucketRoot))
		if err != nil {
			logs.Std().Error(err)
			return nil
		}
		return root
	} else {
		return tx.Bucket(util.StrToBytes(MetadataBucketRoot))
	}
}

func CreateVersionBucket(tx *bolt.Tx, name string) error {
	root := getVersionRoot(tx)
	if root == nil {
		return errors.New("version root is nil")
	}
	if _, err := root.CreateBucket(util.StrToBytes(name)); err != nil {
		return fmt.Errorf("create version bucket: %w", err)
	}
	return nil
}

func RemoveVersionBucket(tx *bolt.Tx, name string) error {
	return getVersionRoot(tx).DeleteBucket(util.StrToBytes(name))
}

// getVersionRoot get or create version root bucket
func getVersionRoot(tx *bolt.Tx) *bolt.Bucket {
	if tx.Writable() {
		root, err := tx.CreateBucketIfNotExists(util.StrToBytes(VersionBucketRoot))
		if err != nil {
			logs.Std().Error(err)
			return nil
		}
		return root
	} else {
		return tx.Bucket(util.StrToBytes(VersionBucketRoot))
	}
}

// GetVersionBucket get version bucket for given name
func GetVersionBucket(tx *bolt.Tx, name string) *bolt.Bucket {
	if root := getVersionRoot(tx); root != nil {
		return root.Bucket(util.StrToBytes(name))
	}
	return nil
}

func getVer(bucket *bolt.Bucket, name string, ver uint64, dest *entity.Version) error {
	if bucket == nil {
		return ErrNotFound
	}
	bt := bucket.Get(util.StrToBytes(fmt.Sprint(name, Sep, ver)))
	if bt == nil {
		return ErrNotFound
	}
	return util.DecodeMsgp(dest, bt)
}

func getMeta(b *bolt.Bucket, name string, dest *entity.Metadata) error {
	if b == nil {
		return ErrNotFound
	}
	bt := b.Get(util.StrToBytes(name))
	if bt == nil {
		return ErrNotFound
	}
	return util.DecodeMsgp(dest, bt)
}
