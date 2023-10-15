package logic

import (
	"bytes"
	"common/logs"
	"common/proto/msg"
	"common/util"
	"errors"
	"fmt"
	. "metaserver/internal/usecase"

	"github.com/google/uuid"
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

func AddMeta(id string, data *msg.Metadata) TxFunc {
	return func(tx *bolt.Tx) error {
		root := GetMetadataBucket(tx)
		key := util.StrToBytes(id)
		// check duplicate
		if root.Get(key) != nil {
			return ErrExists
		}
		// not save Extra
		data.Extra = nil
		// encode data
		bt, err := util.EncodeMsgp(data)
		if err != nil {
			return err
		}
		// create version bucket
		if err = CreateVersionBucket(tx, id); err != nil {
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
			return nil
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

func UpdateMeta(id string, data *msg.Metadata) TxFunc {
	return func(tx *bolt.Tx) error {
		root := GetMetadataBucket(tx)
		var origin msg.Metadata
		if err := getMeta(root, id, &origin); err != nil {
			return err
		}
		if data.UpdateTime <= origin.UpdateTime {
			return ErrOldData
		}
		// not save Extra
		data.Extra = nil
		data.Bucket = origin.Bucket
		data.CreateTime = origin.CreateTime
		bt, err := util.EncodeMsgp(data)
		if err != nil {
			return err
		}
		return root.Put(util.StrToBytes(id), bt)
	}
}

func GetMeta(name string, data *msg.Metadata) TxFunc {
	return func(tx *bolt.Tx) error {
		return getMeta(GetMetadataBucket(tx), name, data)
	}
}

func GetExtra(id string, extra *msg.Extra) TxFunc {
	return func(tx *bolt.Tx) error {
		b := GetVersionBucket(tx, id)
		if b == nil {
			return ErrNotFound
		}
		extra.Total = b.Stats().KeyN
		cur := b.Cursor()
		// first key
		k, _ := cur.First()
		if idx := bytes.LastIndexByte(k, Sep[0]); idx > 0 {
			extra.FirstVersion = util.ToInt(util.BytesToStr(k[idx+1:]))
		}
		// last key
		k, _ = cur.Last()
		if idx := bytes.LastIndexByte(k, Sep[0]); idx > 0 {
			extra.LastVersion = util.ToInt(util.BytesToStr(k[idx+1:]))
		}
		return nil
	}
}

func AddVerWithSequence(name string, data *msg.Version) TxFunc {
	return func(tx *bolt.Tx) error {
		if bucket := GetVersionBucket(tx, name); bucket != nil {
			var byUniqueId []string
			if err := NewUniqueHashIndex().GetIndex(data.UniqueId, &byUniqueId)(tx); err != nil {
				return err
			}
			if (len(byUniqueId) > 0) {
				return ErrExists
			}
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
			bt, err := util.EncodeMsgp(data)
			if err != nil {
				return err
			}
			if err := bucket.Put(key, bt); err != nil {
				return err
			}
			return NewHashIndexLogic().AddIndex(data.Hash, util.BytesToStr(key))(tx)
		}
		return ErrNotFound
	}
}

func AddVer(name string, data *msg.Version) TxFunc {
	return func(tx *bolt.Tx) error {
		if bucket := GetVersionBucket(tx, name); bucket != nil {
			var byUniqueId []string
			if err := NewUniqueHashIndex().GetIndex(data.UniqueId, &byUniqueId)(tx); err != nil {
				return err
			}
			if (len(byUniqueId) > 0) {
				return ErrExists
			}
			data.Sequence, _ = bucket.NextSequence()
			keyStr := fmt.Sprint(name, Sep, data.Sequence)
			key := util.StrToBytes(keyStr)
			if bucket.Get(key) != nil {
				return ErrExists
			}
			bt, err := util.EncodeMsgp(data)
			if err != nil {
				return err
			}
			if err = bucket.Put(key, bt); err != nil {
				return err
			}
			if err = NewHashIndexLogic().AddIndex(data.Hash, keyStr)(tx); err != nil {
				return err
			}
			return NewUniqueHashIndex().AddIndex(data.UniqueId, keyStr)(tx)
		}
		return ErrNotFound
	}
}

func RemoveVer(name string, ver uint64) TxFunc {
	return func(tx *bolt.Tx) error {
		key := util.StrToBytes(fmt.Sprint(name, Sep, ver))
		b := GetVersionBucket(tx, name)
		if b == nil {
			return nil
		}
		var data msg.Version
		if err := getVer(b, name, ver, &data); err != nil {
			return err
		}
		// remove index
		if err := NewHashIndexLogic().RemoveIndex(data.Hash, util.BytesToStr(key))(tx); err != nil {
			return fmt.Errorf("remove hash-index err: %w", err)
		}
		return b.Delete(key)
	}
}

func UpdateVer(id string, data *msg.Version) TxFunc {
	return func(tx *bolt.Tx) error {
		if b := GetVersionBucket(tx, id); b != nil {
			key := util.StrToBytes(fmt.Sprint(id, Sep, data.Sequence))
			// get old one
			var origin msg.Version
			if err := getVer(b, id, data.Sequence, &origin); err != nil {
				return err
			}
			// validate timestamp
			if data.Ts <= origin.Ts {
				return ErrOldData
			}
			// those updating are not allowed
			data.Sequence = origin.Sequence
			data.Hash = origin.Hash
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

func GetVer(id string, ver uint64, dest *msg.Version) TxFunc {
	return func(tx *bolt.Tx) error {
		if bucket := GetVersionBucket(tx, id); bucket != nil {
			return getVer(bucket, id, ver, dest)
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
		root.FillPercent = 0.9
		return root
	} else {
		b := tx.Bucket(util.StrToBytes(MetadataBucketRoot))
		if b == nil {
			return b
		}
		b.FillPercent = 0.9
		return b
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
		if b := root.Bucket(util.StrToBytes(name)); b != nil {
			b.FillPercent = 0.9
			return b
		}
	}
	return nil
}

func getVer(bucket *bolt.Bucket, id string, ver uint64, dest *msg.Version) error {
	if bucket == nil {
		return ErrNotFound
	}
	bt := bucket.Get(util.StrToBytes(fmt.Sprint(id, Sep, ver)))
	if bt == nil {
		return ErrNotFound
	}
	return util.DecodeMsgp(dest, bt)
}

func getMeta(b *bolt.Bucket, name string, dest *msg.Metadata) error {
	if b == nil {
		return ErrNotFound
	}
	bt := b.Get(util.StrToBytes(name))
	if bt == nil {
		return ErrNotFound
	}
	return util.DecodeMsgp(dest, bt)
}

// GenerateUniqueId generate an unique id by UUID
// In a single writable cluster, uuid is safe
func GenerateUniqueId() string {
	return uuid.NewString()
}