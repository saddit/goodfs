package service

import (
	"metaserver/internal/entity"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/raftimpl"
	"time"
)

type BucketService struct {
	usecase.BucketRepo
	usecase.RaftApply
}

func NewBucketService(repo usecase.BucketRepo, rw *raftimpl.RaftWrapper) *BucketService {
	return &BucketService{BucketRepo: repo, RaftApply: raftimpl.RaftApplier(rw)}
}

func (b *BucketService) Create(bucket *entity.Bucket) error {
	if bucket == nil {
		return usecase.ErrNilData
	}
	bucket.CreateTime = time.Now().UnixMilli()
	bucket.UpdateTime = bucket.CreateTime
	if ok, resp := b.ApplyRaft(&entity.RaftData{
		Type:   entity.LogInsert,
		Dest:   entity.DestBucket,
		Name:   bucket.Name,
		Bucket: bucket,
	}); ok {
		if resp.Ok() {
			return nil
		}
		return resp
	}
	return b.BucketRepo.Create(bucket)
}

func (b *BucketService) Remove(name string) error {
	if ok, resp := b.ApplyRaft(&entity.RaftData{
		Type: entity.LogRemove,
		Dest: entity.DestBucket,
		Name: name,
	}); ok {
		if resp.Ok() {
			return nil
		}
		return resp
	}
	return b.BucketRepo.Remove(name)
}

func (b *BucketService) Update(bucket *entity.Bucket) error {
	if bucket == nil {
		return usecase.ErrNilData
	}
	bucket.UpdateTime = time.Now().UnixMilli()
	if ok, resp := b.ApplyRaft(&entity.RaftData{
		Type:   entity.LogUpdate,
		Dest:   entity.DestBucket,
		Name:   bucket.Name,
		Bucket: bucket,
	}); ok {
		if resp.Ok() {
			return nil
		}
		return resp
	}
	return b.BucketRepo.Update(bucket)
}
