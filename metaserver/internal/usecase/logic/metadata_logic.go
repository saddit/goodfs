package logic

import (
	"fmt"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
	"metaserver/internal/usecase/utils"
	"time"

	bolt "go.etcd.io/bbolt"
)

const (
	BucketName = "goodfs.metadata"
	Sep        = "."
)

func AddMeta(name string, data *entity.Metadata) TxFunc {
	return func(tx *bolt.Tx) error {
		root := GetRoot(tx)
		key := []byte(name)
		// check duplicate
		if root.Get(key) != nil {
			return ErrExists
		}
		// encode data
		bt := utils.EncodeMsgp(data)
		if bt == nil {
			return ErrEncode
		}
		data.CreateTime = time.Now().Unix()
		data.UpdateTime = time.Now().Unix()
		data.Name = name
		// create version bucket
		if _, err := root.CreateBucket([]byte("nest_" + name)); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		// put metadata
		return root.Put(key, bt)
	}
}

func RemoveMeta(name string) TxFunc {
	return func(tx *bolt.Tx) error {
		key := []byte(name)
		root := GetRoot(tx)
		if root.Get(key) == nil {
			return ErrNotFound
		}
		if err := root.Delete(key); err != nil {
			return err
		}
		return root.DeleteBucket(key)
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
		if bt := utils.EncodeMsgp(data); bt != nil {
			// update data
			data.UpdateTime = time.Now().Unix()
			return root.Put([]byte(name), bt)
		}
		return ErrEncode
	}
}

func GetMeta(name string, data *entity.Metadata) TxFunc {
	return func(tx *bolt.Tx) error {
		return getMeta(GetRoot(tx), name, data)
	}
}

func AddVer(name string, data *entity.Version) TxFunc {
	return func(tx *bolt.Tx) error {
		if bucket := GetRootNest(tx, name); bucket != nil {
			data.Sequence, _ = bucket.NextSequence()
			data.Ts = time.Now().Unix()
			key := []byte(fmt.Sprint(name, Sep, data.Sequence))
			if bt := utils.EncodeMsgp(data); bt != nil {
				return bucket.Put(key, bt)
			}
			return ErrEncode
		}
		return ErrNotFound
	}
}

func RemoveVer(name string, ver int) TxFunc {
	return func(tx *bolt.Tx) error {
		key := []byte(fmt.Sprint(name, Sep, ver))
		b := GetRootNest(tx, name)
		if b != nil {
			return ErrNotFound
		}
		if err := b.Delete(key); err != nil {
			return ErrNotFound
		}
		return nil
	}
}

func UpdateVer(name string, data *entity.Version) TxFunc {
	return func(tx *bolt.Tx) error {
		if b := GetRootNest(tx, name); b != nil {
			key := []byte(fmt.Sprint(name, Sep, data.Sequence))
			// validate ts
			var origin entity.Version
			if err := getVer(b, name, data.Sequence, &origin); err != nil {
				return err
			}
			if data.Ts < origin.Ts {
				return ErrOldData
			}
			if bt := utils.EncodeMsgp(data); bt != nil {
				// update data
				data.Ts = time.Now().Unix()
				return b.Put(key, bt)
			}
			return ErrEncode
		}
		return ErrNotFound
	}
}

func GetVer(name string, ver uint64, dest *entity.Version) TxFunc {
	return func(tx *bolt.Tx) error {
		if bucket := GetRootNest(tx, name); bucket != nil {
			getVer(bucket, name, ver, dest)
			return nil
		}
		return ErrNotFound
	}
}

func GetRoot(tx *bolt.Tx) *bolt.Bucket {
	if tx.Writable() {
		root, err := tx.CreateBucketIfNotExists([]byte(BucketName))
		if err != nil {
			panic(err)
		}
		return root
	} else {
		return tx.Bucket([]byte(BucketName))
	}
}

func GetRootNest(tx *bolt.Tx, name string) *bolt.Bucket {
	return GetRoot(tx).Bucket([]byte("nest_" + name))
}

func getVer(bucket *bolt.Bucket, name string, ver uint64, dest *entity.Version) error {
	if bucket == nil {
		return ErrNotFound
	}
	bt := bucket.Get([]byte(fmt.Sprint(name, Sep, ver)))
	if bt == nil {
		return ErrNotFound
	}
	if !utils.DecodeMsgp(dest, bt) {
		return ErrDecode
	}
	return nil
}

func getMeta(b *bolt.Bucket, name string, dest *entity.Metadata) error {
	if b == nil {
		return ErrNotFound
	}
	bt := b.Get([]byte(name))
	if bt == nil {
		return ErrNotFound
	}
	if !utils.DecodeMsgp(dest, bt) {
		return ErrDecode
	}
	return nil
}
