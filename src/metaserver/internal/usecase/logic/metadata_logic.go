package logic

import (
	"errors"
	"fmt"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
	"strings"
	"time"

	"common/util"

	bolt "go.etcd.io/bbolt"
)

const (
	RootBucketName = "goodfs.metadata"
	NestPrefix     = "nest_"
	Sep            = "."
)

func ForeachKeys(fn func(string) bool) TxFunc {
	return func(tx *bolt.Tx) error {
		root := GetRoot(tx)
		return root.ForEach(func(k, v []byte) error {
			// skip nest bucket keys
			if strings.HasPrefix(string(k), NestPrefix) {
				return nil
			}
			if !fn(string(k)) {
				return ErrNotFound
			}
			return nil
		})
	}
}

func AddMeta(data *entity.Metadata) TxFunc {
	return func(tx *bolt.Tx) error {
		root := GetRoot(tx)
		key := util.StrToBytes(data.Name)
		// check duplicate
		if root.Get(key) != nil {
			return ErrExists
		}
		// encode data
		bt, err := util.EncodeMsgp(data)
		if err != nil {
			return err
		}
		data.CreateTime = time.Now().UnixMilli()
		data.UpdateTime = data.CreateTime
		// create version bucket
		if _, err := root.CreateBucket(util.StrToBytes(NestPrefix + data.Name)); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		// put metadata
		return root.Put(key, bt)
	}
}

func RemoveMeta(name string) TxFunc {
	return func(tx *bolt.Tx) error {
		key := util.StrToBytes(name)
		root := GetRoot(tx)
		if root.Get(key) == nil {
			return ErrNotFound
		}
		if err := root.Delete(key); err != nil {
			return err
		}
		err := root.DeleteBucket(util.StrToBytes(fmt.Sprint(NestPrefix, key)))
		// ignore err of bucket not found
		if err != nil && !errors.Is(err, bolt.ErrBucketNotFound) {
			return err
		}
		return nil
	}
}

func UpdateMeta(name string, data *entity.Metadata) TxFunc {
	return func(tx *bolt.Tx) error {
		root := GetRoot(tx)
		var origin entity.Metadata
		if err := getMeta(root, name, &origin); err != nil {
			return err
		}
		if data.UpdateTime < origin.UpdateTime {
			return ErrOldData
		}
		bt, err := util.EncodeMsgp(data)
		if err != nil {
			return err
		}
		// update data
		data.UpdateTime = time.Now().UnixMilli()
		return root.Put(util.StrToBytes(name), bt)
	}
}

func GetMeta(name string, data *entity.Metadata) TxFunc {
	return func(tx *bolt.Tx) error {
		return getMeta(GetRoot(tx), name, data)
	}
}

func AddVerWithSequence(name string, data *entity.Version) TxFunc {
	return func(tx *bolt.Tx) error {
		if bucket := GetRootNest(tx, name); bucket != nil {
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
		if bucket := GetRootNest(tx, name); bucket != nil {
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
		b := GetRootNest(tx, name)
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
		if b := GetRootNest(tx, name); b != nil {
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
		if bucket := GetRootNest(tx, name); bucket != nil {
			return getVer(bucket, name, ver, dest)
		}
		return ErrNotFound
	}
}

func GetRoot(tx *bolt.Tx) *bolt.Bucket {
	if tx.Writable() {
		root, err := tx.CreateBucketIfNotExists(util.StrToBytes(RootBucketName))
		if err != nil {
			panic(err)
		}
		return root
	} else {
		return tx.Bucket(util.StrToBytes(RootBucketName))
	}
}

func GetRootNest(tx *bolt.Tx, name string) *bolt.Bucket {
	if root := GetRoot(tx); root != nil {
		return root.Bucket(util.StrToBytes(NestPrefix + name))
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
