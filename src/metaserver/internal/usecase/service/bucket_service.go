package service

import (
	"metaserver/internal/entity"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/pool"
	"metaserver/internal/usecase/raftimpl"
)

type BucketService struct {
	usecase.BucketRepo
	usecase.RaftApply
}

func NewBucketService(repo usecase.BucketRepo) *BucketService {
	return &BucketService{BucketRepo: repo, RaftApply: raftimpl.RaftApplier(pool.RaftWrapper)}
}

func (b *BucketService) Create(bucket *entity.Bucket) error {
	if bucket == nil {
		return usecase.ErrNilData
	}
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
