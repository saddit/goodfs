package logic

import (
	"common/util"
	"metaserver/internal/usecase"

	bolt "go.etcd.io/bbolt"
)

const (
	UniqueIdIndexName = "uniqueIdIndex"
)

type UniqueIdIndex struct{}

func NewUniqueIdIndex() UniqueIdIndex {
	return UniqueIdIndex{}
}

func (UniqueIdIndex) AddIndex(uniqueId, key string) usecase.TxFunc {
	return func(tx *bolt.Tx) error {
		buk := GetIndexBucket(tx, UniqueIdIndexName)
		uniqueIdBuk, err := buk.CreateBucketIfNotExists(util.StrToBytes(uniqueId))
		if err != nil {
			return err
		}
		return uniqueIdBuk.Put(util.StrToBytes(key), []byte{})
	}
}

func (UniqueIdIndex) RemoveIndex(uniqueId, key string) usecase.TxFunc {
	return func(tx *bolt.Tx) error {
		buk := GetIndexBucket(tx, UniqueIdIndexName)
		if uniqueIdBuk := buk.Bucket(util.StrToBytes(uniqueId)); uniqueIdBuk != nil {
			return uniqueIdBuk.Delete(util.StrToBytes(key))
		}
		return nil
	}
}

func (UniqueIdIndex) GetIndex(uniqueId string, res *[]string) usecase.TxFunc {
	*res = []string{}
	return func(tx *bolt.Tx) error {
		buk := GetIndexBucket(tx, UniqueIdIndexName)
		if buk == nil {
			return nil
		}
		uniqueIdBuk := buk.Bucket(util.StrToBytes(uniqueId))
		if uniqueIdBuk == nil {
			return nil
		}
		err := uniqueIdBuk.ForEach(func(k, v []byte) error {
			*res = append(*res, util.BytesToStr(k))
			return nil
		})
		return err
	}
}
