package repo

import (
	"common/proto/msg"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/logic"
)

type BatchBucketRepo struct {
	db    *db.Storage
	logic *logic.BucketCrud
}

func NewBatchBucketRepo(db *db.Storage) *BatchBucketRepo {
	return &BatchBucketRepo{db: db, logic: logic.NewBucketCrud()}
}

func (b *BatchBucketRepo) Create(bucket *msg.Bucket) (err error) {
	return b.db.Batch(b.logic.Create(bucket))
}

func (b *BatchBucketRepo) Remove(name string) (err error) {
	return b.db.Batch(b.logic.Delete(name))
}

func (b *BatchBucketRepo) Update(bucket *msg.Bucket) (err error) {
	return b.db.Batch(b.logic.Update(bucket))
}

func (b *BatchBucketRepo) Sync() error {
	return b.db.DB().Sync()
}
