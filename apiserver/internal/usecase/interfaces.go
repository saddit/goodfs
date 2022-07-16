package usecase

import (
	"apiserver/internal/entity"
	"io"
)

type (
	IMetaService interface {
		SaveMetadata(*entity.Metadata) (int32, error)
		UpdateVersion(*entity.Version)
		GetVersion(string, int32) (*entity.Version, bool)
		GetMetadata(string, int32) (*entity.Metadata, bool)
	}
	IObjectService interface {
		LocateObject(hash string) ([]string, bool)
		StoreObject(req *entity.PutReq, md *entity.Metadata) (int32, error)
		GetObject(ver *entity.Version) (io.ReadSeekCloser, error)
	}
)
