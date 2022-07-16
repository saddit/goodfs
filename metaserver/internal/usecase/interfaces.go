package usecase

import "metaserver/internal/entity"

type (
	//IMetadataService 负责格式转换，缓存处理等
	IMetadataService interface {
		AddMetadata(string, *entity.Metadata) error
		AddVersion(string, *entity.Version) error
		UpdateMetadata(string, *entity.Metadata) error
		UpdateVersion(string, int, *entity.Version) error
		RemoveMetadata(string) error
		RemoveVersion(string, int) error
		GetMetadata(string) (*entity.Metadata, error)
		GetVersion(string, int) (*entity.Metadata, error)
		ListVersions(string, int, int) ([]*entity.Version, error)
	}

	//IMetadataRepo 负责对文件系统存储
	IMetadataRepo interface {
		ExistMetadata(string) bool
		AddMetadata(string, *entity.Metadata) error
		AddVersion(string, *entity.Version) (int, error)
		UpdateMetadata(string, *entity.Metadata) error
		UpdateVersion(string, *entity.Version) error
		RemoveMetadata(string) error
		RemoveVersion(string, int) error
		GetMetadata(string) (*entity.Metadata, error)
		GetVersion(string, int) (*entity.Version, error)
		ListVersions(string, int, int) ([]*entity.Version, error)
	}
)
