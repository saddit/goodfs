package usecase

import (
	"apiserver/internal/entity"
	"io"
)

type (
	IMetaService interface {
		SaveMetadata(*entity.Metadata) (int32, error)
		UpdateVersion(string, *entity.Version) error
		GetVersion(string, int32) (*entity.Version, error)
		GetMetadata(string, int32) (*entity.Metadata, error)
	}
	IObjectService interface {
		LocateObject(hash string) ([]string, bool)
		StoreObject(req *entity.PutReq, md *entity.Metadata) (int32, error)
		GetObject(meta *entity.Metadata, ver *entity.Version) (io.ReadSeekCloser, error)
	}
)
