package logic

import (
	"common/util"
	"fmt"
	"metaserver/internal/usecase"

	bolt "go.etcd.io/bbolt"
)

const (
	HashIndexName = "hashIndex"
)

type HashIndexLogic struct{}

func NewHashIndexLogic() HashIndexLogic { return HashIndexLogic{} }

func (HashIndexLogic) AddIndex(hash, key string) usecase.TxFunc {
	return func(tx *bolt.Tx) error {
		buk := GetIndexBucket(tx, HashIndexName)
		hashBuk, err := buk.CreateBucketIfNotExists(util.StrToBytes(hash))
		if err != nil {
			return err
		}
		return hashBuk.Put(util.StrToBytes(key), []byte{})
	}
}

func (HashIndexLogic) RemoveIndex(hash, key string) usecase.TxFunc {
	return func(tx *bolt.Tx) error {
		buk := GetIndexBucket(tx, HashIndexName)
		if hashBuk := buk.Bucket(util.StrToBytes(hash)); hashBuk != nil {
			return hashBuk.Delete(util.StrToBytes(key))
		}
		return nil
	}
}

func (HashIndexLogic) GetIndex(hash string, res *[]string) usecase.TxFunc {
	*res = []string{}
	return func(tx *bolt.Tx) error {
		buk := GetIndexBucket(tx, HashIndexName)
		if buk == nil {

			return nil
		}
		hashBuk := buk.Bucket(util.StrToBytes(hash))
		if hashBuk == nil {
			return nil
		}
		err := hashBuk.ForEach(func(k, v []byte) error {
			*res = append(*res, util.BytesToStr(k))
			return nil
		})
		return err
	}
}

func GetIndexBucket(tx *bolt.Tx, indexName string) *bolt.Bucket {
	bt := util.StrToBytes(fmt.Sprint("goodfs.metadata.", indexName))
	if tx.Writable() {
		res, _ := tx.CreateBucketIfNotExists(bt)
		return res
	}
	return tx.Bucket(bt)
}
