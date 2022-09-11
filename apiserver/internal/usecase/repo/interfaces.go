package repo

import (
	"apiserver/internal/entity"
)

type IMetadataRepo interface {
	FindByName(name string) (*entity.Metadata, error)
	FindByNameWithVersion(name string, verMode entity.VerMode) (*entity.Metadata, error)
	Insert(data *entity.Metadata) (*entity.Metadata, error)
}

type IVersionRepo interface {
	Find(string, int32) (*entity.Version, error)
	Update(name string, ver *entity.Version) error
	Add(name string, ver *entity.Version) (int32, error)
	Delete(name string, ver *entity.Version) error
}
