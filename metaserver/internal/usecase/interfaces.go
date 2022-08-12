package usecase

import (
	"metaserver/internal/entity"
	"time"
	"github.com/hashicorp/raft"
	bolt "go.etcd.io/bbolt"
)

type (
	//IMetadataService 负责格式转换，缓存处理等
	IMetadataService interface {
		AddMetadata(string, *entity.Metadata) error
		AddVersion(string, *entity.Version) (int, error)
		UpdateMetadata(string, *entity.Metadata) error
		UpdateVersion(string, int, *entity.Version) error
		RemoveMetadata(string) error
		RemoveVersion(string, int) error
		GetMetadata(string, int) (*entity.Metadata, *entity.Version, error)
		GetVersion(string, int) (*entity.Version, error)
		ListVersions(string, int, int) ([]*entity.Version, error)
	}

	//IMetadataRepo 负责对文件系统存储
	IMetadataRepo interface {
		AddMetadata(string, *entity.Metadata) error
		AddVersion(string, *entity.Version) error
		UpdateMetadata(string, *entity.Metadata) error
		UpdateVersion(string, *entity.Version) error
		RemoveMetadata(string) error
		RemoveVersion(string, int) error
		GetMetadata(string) (*entity.Metadata, error)
		GetVersion(string, uint64) (*entity.Version, error)
		ListVersions(string, int, int) ([]*entity.Version, error)
		ApplyRaft(*entity.RaftData) error
	}

	TxFunc func(*bolt.Tx) error

	ITransaction interface {
		Update(func(*bolt.Tx) error) error
		Batch(func(*bolt.Tx) error) error
		View(func(*bolt.Tx) error) error
	}

	IRaft interface {
		Apply(cmd []byte, timeout time.Duration) raft.ApplyFuture
		ApplyLog(log raft.Log, timeout time.Duration) raft.ApplyFuture
	}
)
