package repo

import (
	"apiserver/internal/entity"
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

type IMetadataRepo interface {
	Find(filter bson.M, verMode entity.VerMode) (*entity.MetaData, error)
	FindById(id string) *entity.MetaData
	FindByName(name string) *entity.MetaData
	FindByNameAndVerMode(name string, verMode entity.VerMode) *entity.MetaData
	FindByHash(hash string) *entity.MetaData
	Insert(data *entity.MetaData) (*entity.MetaData, error)
	Exist(filter bson.M) bool
	Update(data *entity.MetaData) error
}

type IVersionRepo interface {
	Find(hash string) (*entity.Version, int32)
	Update(ctx context.Context, ver *entity.Version) bool
	Add(ctx context.Context, id string, ver *entity.Version) int32
	Delete(ctx context.Context, id string, ver *entity.Version) error
}
