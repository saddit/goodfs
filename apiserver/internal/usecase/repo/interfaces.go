package repo

import (
	"apiserver/internal/entity"
	"context"
)

type IMetadataRepo interface {
	// Find(filter bson.M, verMode entity.VerMode) (*entity.MetaData, error)
	// FindById(id string) *entity.MetaData
	// FindByHash(hash string) *entity.MetaData
	// Exist(filter bson.M) bool
	// Update(data *entity.MetaData) error
	FindByName(name string) *entity.Metadata
	FindByNameAndVerMode(name string, verMode entity.VerMode) *entity.Metadata
	Insert(data *entity.Metadata) (*entity.Metadata, error)
}

type IVersionRepo interface {
	Find(string, int32) *entity.Version
	Update(ctx context.Context, ver *entity.Version) bool
	Add(ctx context.Context, id string, ver *entity.Version) int32
	Delete(ctx context.Context, id string, ver *entity.Version) error
}
