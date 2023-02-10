package repo

import (
	"common/cache"
	"common/proto/msg"
	"common/util"
	"fmt"
	"metaserver/internal/usecase"
)

const (
	BucketCachePrefix = "bucket_"
)

type BucketCacheRepo struct {
	cache cache.ICache
}

func NewBucketCacheRepo(cache cache.ICache) *BucketCacheRepo {
	return &BucketCacheRepo{cache: cache}
}

func (b *BucketCacheRepo) Get(name string) (*msg.Bucket, error) {
	bt, ok := b.cache.HasGet(fmt.Sprint(BucketCachePrefix, name))
	if !ok {
		return nil, usecase.ErrNotFound
	}
	var i msg.Bucket
	err := util.DecodeMsgp(&i, bt)
	return &i, err
}

func (b *BucketCacheRepo) Create(bucket *msg.Bucket) error {
	bt, err := util.EncodeMsgp(bucket)
	if err != nil {
		return err
	}
	b.cache.Set(fmt.Sprint(BucketCachePrefix, bucket.Name), bt)
	return nil
}

func (b *BucketCacheRepo) Remove(name string) error {
	b.cache.Delete(fmt.Sprint(BucketCachePrefix, name))
	return nil
}

func (b *BucketCacheRepo) Update(bucket *msg.Bucket) error {
	return b.Create(bucket)
}

func (b *BucketCacheRepo) GetBytes(name string) ([]byte, error) {
	bt, ok := b.cache.HasGet(fmt.Sprint(BucketCachePrefix, name))
	if !ok {
		return nil, usecase.ErrNotFound
	}
	return bt, nil
}

func (b *BucketCacheRepo) List(string, int) ([]*msg.Bucket, int, error) {
	panic("not implement Foreach")
}

func (b *BucketCacheRepo) Foreach(func(k []byte, v []byte) error) error {
	panic("not implement Foreach")
}
