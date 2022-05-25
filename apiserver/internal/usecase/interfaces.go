package usecase

import (
	"apiserver/internal/entity"
	"io"
)

type (
	IMetaService interface {
		SaveMetadata(*entity.MetaData) (int32, error)
		UpdateVersion(*entity.Version)
		GetVersion(string) (*entity.Version, int32, bool)
		GetMetadata(string, int32) (*entity.MetaData, bool)
	}
	IObjectService interface {
		LocateObject(hash string) ([]string, bool)
		StoreObject(req *entity.PutReq, md *entity.MetaData) (int32, error)
		GetObject(ver *entity.Version) (io.ReadSeekCloser, error)
	}
)
