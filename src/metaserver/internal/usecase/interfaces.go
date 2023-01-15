package usecase

import (
	"common/pb"
	"common/response"
	"io"
	"metaserver/internal/entity"
	"time"

	"github.com/hashicorp/raft"
	bolt "go.etcd.io/bbolt"
)

type (
	MetadataRpcService interface {
		ForeachVersionBytes(string, func([]byte) bool)
		GetMetadataBytes(string) ([]byte, error)
		FilterKeys(fn func(string) bool) []string
		FindByHash(hash string) (res []*pb.Version, err error)
		UpdateLocates(name string, seq int, locates []string) error
	}

	IMetadataService interface {
		MetadataRpcService
		ReceiveVersion(string, *entity.Version) error
		AddMetadata(*entity.Metadata) error
		AddVersion(string, *entity.Version) (int, error)
		UpdateMetadata(string, *entity.Metadata) error
		UpdateVersion(string, int, *entity.Version) error
		RemoveMetadata(string) error
		RemoveVersion(string, int) error
		GetMetadata(string, int) (*entity.Metadata, *entity.Version, error)
		GetVersion(string, int) (*entity.Version, error)
		ListVersions(string, int, int) ([]*entity.Version, int, error)
		ListMetadata(prefix string, size int) (lst []*entity.Metadata, total int, err error)
	}

	WritableRepo interface {
		AddMetadata(*entity.Metadata) error
		AddVersion(string, *entity.Version) error
		UpdateMetadata(string, *entity.Metadata) error
		UpdateVersion(string, *entity.Version) error
		RemoveMetadata(string) error
		RemoveVersion(string, uint64) error
	}

	ReadableRepo interface {
		GetMetadata(string) (*entity.Metadata, error)
		GetVersion(string, uint64) (*entity.Version, error)
		ListVersions(string, int, int) ([]*entity.Version, int, error)
		ListMetadata(prefix string, size int) (lst []*entity.Metadata, total int, err error)
	}

	IHashIndexRepo interface {
		Remove(hash, key string) error
		FindAll(hash string) ([]string, error)
		Sync() error
	}

	IBatchMetaRepo interface {
		WritableRepo
		ForeachKeys(func(string) bool)
		Sync() error
	}

	SnapshotTx interface {
		io.WriterTo
		Rollback() error
	}

	SnapshotManager interface {
		Snapshot() (SnapshotTx, error)
		Restore(io.Reader) error
	}

	RaftApply interface {
		ApplyRaft(*entity.RaftData) (bool, *response.RaftFsmResp)
	}

	IMetadataRepo interface {
		WritableRepo
		ReadableRepo
		SnapshotManager
		AddVersionWithSequence(string, *entity.Version) error
		RemoveAllVersion(string) error
		GetLastVersionNumber(name string) uint64
		GetFirstVersionNumber(name string) uint64
		ForeachVersionBytes(string, func([]byte) bool)
		GetMetadataBytes(string) ([]byte, error)
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

	IRaftLeaderChanged interface {
		OnLeaderChanged(bool)
	}

	IHashSlotService interface {
		AutoMigrate(toLoc *pb.LocationInfo, slots []string) error
		PrepareMigrationFrom(loc *pb.LocationInfo, slots []string) error
		PrepareMigrationTo(loc *pb.LocationInfo, slots []string) error
		ReceiveItem(*pb.MigrationItem) error
		FinishReceiveItem(bool) error
		GetCurrentSlots(bool) (map[string][]string, error)
	}

	IMetaCache interface {
		ReadableRepo
		WritableRepo
	}

	BucketRepo interface {
		Foreach(func(k []byte, v []byte) error) error
		Get(name string) (*entity.Bucket, error)
		GetBytes(name string) ([]byte, error)
		Create(bucket *entity.Bucket) error
		Remove(name string) error
		Update(bucket *entity.Bucket) error
		List(prefix string, size int) ([]*entity.Bucket, int, error)
	}

	BucketService interface {
		BucketRepo
	}
)
