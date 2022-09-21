package logic

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

const (
	HashIndexName = "hashIndex"
)

type HashIndexLogic struct {}

func NewHashIndexLogic() HashIndexLogic {return HashIndexLogic{}}

func (HashIndexLogic) AddIndex(tx *bolt.Tx, hash, key string) error {
	buk := GetIndexBucket(tx, HashIndexName)
	hashBuk, err := buk.CreateBucketIfNotExists([]byte(hash))
	if err != nil {
		return err
	}
	return hashBuk.Put([]byte{}, []byte{})
}

func (HashIndexLogic) RemoveIndex(tx *bolt.Tx, hash, key string) error {
	buk := GetIndexBucket(tx, HashIndexName)
	if hashBuk := buk.Bucket([]byte(hash)); hashBuk != nil {
		return hashBuk.Delete([]byte(key))
	}
	return nil
}

func (HashIndexLogic) GetIndex(tx *bolt.Tx, hash string) ([]string, error) {
	buk := GetIndexBucket(tx, HashIndexName)
	var keys []string
	if buk == nil {
		return keys, nil
	}
	hashBuk := buk.Bucket([]byte(hash))
	if hashBuk == nil {
		return keys, nil
	}
	err := hashBuk.ForEach(func(k, v []byte) error {
		keys = append(keys, string(k))
		return nil
	})
	return keys, err
}

func GetIndexBucket(tx *bolt.Tx, indexName string) *bolt.Bucket {
	bt := []byte(fmt.Sprint("goodfs.metadata.", indexName))
	if tx.Writable() {
		res, _ := tx.CreateBucketIfNotExists(bt)
		return res
	}
	return tx.Bucket(bt)
}