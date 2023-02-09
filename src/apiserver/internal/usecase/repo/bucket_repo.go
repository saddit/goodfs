package repo

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/grpcapi"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/webapi"
	"common/response"
)

type BucketRepo struct {
}

func (b *BucketRepo) Get(s string) (*entity.Bucket, error) {
	masterId, err := logic.NewHashSlot().KeySlotLocation(s)
	if err != nil {
		return nil, err
	}
	ip, err := logic.NewDiscovery().SelectMetaServerGRPC(masterId)
	if err != nil {
		return nil, err
	}
	return grpcapi.GetBucket(ip, s)
}

func (b *BucketRepo) Update(bucket *entity.Bucket) error {
	if bucket.Name == "" {
		return response.NewError(400, "bucket name required")
	}
	masterId, err := logic.NewHashSlot().KeySlotLocation(bucket.Name)
	if err != nil {
		return err
	}
	return webapi.PutBucket(logic.NewDiscovery().GetMetaServerHTTP(masterId), bucket)
}

func (b *BucketRepo) Create(bucket *entity.Bucket) error {
	if bucket.Name == "" {
		return response.NewError(400, "bucket name required")
	}
	masterId, err := logic.NewHashSlot().KeySlotLocation(bucket.Name)
	if err != nil {
		return err
	}
	return grpcapi.SaveBucket(logic.NewDiscovery().GetMetaServerGRPC(masterId), bucket)
}

func (b *BucketRepo) Delete(s string) error {
	masterId, err := logic.NewHashSlot().KeySlotLocation(s)
	if err != nil {
		return err
	}
	return webapi.DeleteBucket(logic.NewDiscovery().GetMetaServerHTTP(masterId), s)
}

func NewBucketRepo() *BucketRepo {
	return &BucketRepo{}
}
