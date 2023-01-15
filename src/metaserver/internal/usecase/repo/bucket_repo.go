package repo

import (
	"common/graceful"
	"common/util"
	"metaserver/internal/entity"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/logic"
)

type BucketRepo struct {
	db    *db.Storage
	logic *logic.BucketCrud
	cache usecase.BucketRepo
}

func NewBucketRepo(db *db.Storage, cache usecase.BucketRepo) *BucketRepo {
	return &BucketRepo{db: db, logic: logic.NewBucketCrud(), cache: cache}
}

func (b *BucketRepo) Foreach(fn func(k []byte, v []byte) error) error {
	return b.db.View(b.logic.Foreach(fn))
}

func (b *BucketRepo) Get(name string) (res *entity.Bucket, err error) {
	if data, err := b.cache.Get(name); err == nil {
		return data, nil
	}
	res = new(entity.Bucket)
	err = b.db.View(b.logic.Get(name, res))
	go func() {
		defer graceful.Recover()
		if err == nil {
			util.LogErrWithPre("bucket cache", b.cache.Create(res))
		}
	}()
	return
}

func (b *BucketRepo) Create(bucket *entity.Bucket) (err error) {
	defer func() {
		go func() {
			defer graceful.Recover()
			if err == nil {
				util.LogErrWithPre("bucket cache", b.cache.Create(bucket))
			}
		}()
	}()
	return b.db.Update(b.logic.Create(bucket))
}

func (b *BucketRepo) Remove(name string) (err error) {
	defer func() {
		go func() {
			defer graceful.Recover()
			if err == nil {
				util.LogErrWithPre("bucket cache", b.cache.Remove(name))
			}
		}()
	}()
	return b.db.Update(b.logic.Delete(name))
}

func (b *BucketRepo) Update(bucket *entity.Bucket) (err error) {
	defer func() {
		go func() {
			if err == nil {
				util.LogErrWithPre("bucket cache", b.cache.Update(bucket))
			}
		}()
	}()

	return b.db.Update(b.logic.Update(bucket))
}

func (b *BucketRepo) List(prefix string, size int) ([]*entity.Bucket, int, error) {
	var total int
	list := make([]*entity.Bucket, 0, size)
	err := b.db.View(b.logic.List(prefix, size, &list, &total))
	go func() {
		defer graceful.Recover()
		for _, item := range list {
			util.LogErrWithPre("bucket cache", b.cache.Create(item))
		}
	}()
	return list, total, err
}

func (b *BucketRepo) GetBytes(name string) ([]byte, error) {
	if bt, err := b.cache.GetBytes(name); err == nil {
		return bt, nil
	}
	var bt []byte
	err := b.db.View(b.logic.GetBytes(name, &bt))
	return bt, err
}
