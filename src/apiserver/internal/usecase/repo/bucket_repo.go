package repo

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/logic"
	"apiserver/internal/usecase/webapi"
	"common/response"
)

type BucketRepo struct {
}

func (b *BucketRepo) Get(s string) (*entity.Bucket, error) {
	ip, group, err := logic.NewHashSlot().FindMetaLocByName(s)
	if err != nil {
		return nil, err
	}
	ip = logic.NewDiscovery().SelectMetaByGroupID(group, ip)
	return webapi.GetBucket(ip, s)
}

func (b *BucketRepo) Update(bucket *entity.Bucket) error {
	if bucket.Name == "" {
		return response.NewError(400, "bucket name required")
	}
	ip, _, err := logic.NewHashSlot().FindMetaLocByName(bucket.Name)
	if err != nil {
		return err
	}
	return webapi.PutBucket(ip, bucket)
}

func (b *BucketRepo) Create(bucket *entity.Bucket) error {
	if bucket.Name == "" {
		return response.NewError(400, "bucket name required")
	}
	ip, _, err := logic.NewHashSlot().FindMetaLocByName(bucket.Name)
	if err != nil {
		return err
	}
	return webapi.PostBucket(ip, bucket)
}

func (b *BucketRepo) Delete(s string) error {
	ip, _, err := logic.NewHashSlot().FindMetaLocByName(s)
	if err != nil {
		return err
	}
	return webapi.DeleteBucket(ip, s)
}

func NewBucketRepo() *BucketRepo {
	return &BucketRepo{}
}
