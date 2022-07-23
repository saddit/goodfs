package repo

import (
	"apiserver/internal/entity"
	"context"
)

type IMetadataRepo interface {
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
