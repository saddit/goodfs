package usecase

import (
	"apiserver/internal/entity"
	"io"
)

type (
	IMetaService interface {
		SaveMetadata(data *entity.Metadata) (int32, error)
		AddVersion(name, bucket string, version *entity.Version) (int32, error)
		UpdateVersion(name, bucket string, data *entity.Version) error
		GetVersion(name, bucket string, verMode int32) (*entity.Version, error)
		GetMetadata(name, bucket string, verMode int32, withExtra bool) (*entity.Metadata, error)
		RemoveVersion(name, bucket string, version int32) error
	}
	IObjectService interface {
		LocateObject(hash string) ([]string, bool)
		StoreObject(req *entity.PutReq, md *entity.Metadata) (int32, error)
		GetObject(meta *entity.Metadata, ver *entity.Version) (io.ReadSeekCloser, error)
	}
)
