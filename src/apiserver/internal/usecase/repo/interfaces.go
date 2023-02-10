package repo

import (
	"apiserver/internal/entity"
)

type IMetadataRepo interface {
	FindByName(name string, bucket string, withExtra bool) (*entity.Metadata, error)
	Insert(data *entity.Metadata) error
}

type IVersionRepo interface {
	Find(name, bucket string, i int32) (*entity.Version, error)
	Update(name, bucket string, ver *entity.Version) error
	Add(name, bucket string, ver *entity.Version) (int32, error)
	Delete(name, bucket string, ver int32) error
}

type IBucketRepo interface {
	Get(string) (*entity.Bucket, error)
	Update(*entity.Bucket) error
	Create(*entity.Bucket) error
	Delete(string) error
}
