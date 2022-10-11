package logic

import (
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
		hashBuk, err := buk.CreateBucketIfNotExists([]byte(hash))
		if err != nil {
			return err
		}
		return hashBuk.Put([]byte(key), []byte{})
	}
}

func (HashIndexLogic) RemoveIndex(hash, key string) usecase.TxFunc {
	return func(tx *bolt.Tx) error {
		buk := GetIndexBucket(tx, HashIndexName)
		if hashBuk := buk.Bucket([]byte(hash)); hashBuk != nil {
			return hashBuk.Delete([]byte(key))
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
		hashBuk := buk.Bucket([]byte(hash))
		if hashBuk == nil {
			return nil
		}
		err := hashBuk.ForEach(func(k, v []byte) error {
			*res = append(*res, string(k))
			return nil
		})
		return err
	}
}

func GetIndexBucket(tx *bolt.Tx, indexName string) *bolt.Bucket {
	bt := []byte(fmt.Sprint("goodfs.metadata.", indexName))
	if tx.Writable() {
		res, _ := tx.CreateBucketIfNotExists(bt)
		return res
	}
	return tx.Bucket(bt)
}
