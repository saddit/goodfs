package usecase

import (
	"common/proto/msg"
	"common/proto/pb"
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
		FindByHash(hash string) (res []*msg.Version, err error)
		UpdateLocates(hash string, locateIndex int, locate string) error
	}

	IMetadataService interface {
		MetadataRpcService
		ReceiveVersion(string, *msg.Version) error
		AddMetadata(string, *msg.Metadata) error
		AddVersion(string, *msg.Version) (int, error)
		UpdateMetadata(string, *msg.Metadata) error
		UpdateVersion(string, int, *msg.Version) error
		RemoveMetadata(string) error
		RemoveVersion(string, int) error
		GetMetadata(string, int, bool) (*msg.Metadata, *msg.Version, error)
		GetVersion(string, int) (*msg.Version, error)
		ListVersions(string, int, int) ([]*msg.Version, int, error)
		ListMetadata(prefix string, size int) (lst []*msg.Metadata, total int, err error)
	}

	WritableRepo interface {
		AddMetadata(string, *msg.Metadata) error
		AddVersion(string, *msg.Version) error
		UpdateMetadata(string, *msg.Metadata) error
		UpdateVersion(string, *msg.Version) error
		RemoveMetadata(string) error
		RemoveVersion(string, uint64) error
		AddVersionWithSequence(string, *msg.Version) error
		RemoveAllVersion(string) error
	}

	ReadableRepo interface {
		GetMetadata(string) (*msg.Metadata, error)
		GetVersion(string, uint64) (*msg.Version, error)
		ListVersions(string, int, int) ([]*msg.Version, int, error)
		ListMetadata(prefix string, size int) (lst []*msg.Metadata, total int, err error)
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
		LastAppliedIndex() (uint64, error)
		ApplyIndex(i uint64) error
	}

	RaftApply interface {
		ApplyRaft(*entity.RaftData) (bool, any, error)
	}

	IMetadataRepo interface {
		WritableRepo
		ReadableRepo
		UpdateLocateByHash(hash string, index int, value string) error
		GetLastVersionNumber(id string) uint64
		GetFirstVersionNumber(id string) uint64
		ForeachVersionBytes(string, func([]byte) bool)
		GetMetadataBytes(string) ([]byte, error)
		GetExtra(id string) (*msg.Extra, error)
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

	BucketWritableRepo interface {
		Create(bucket *msg.Bucket) error
		Remove(name string) error
		Update(bucket *msg.Bucket) error
	}

	BatchBucketRepo interface {
		BucketWritableRepo
		Sync() error
	}

	BucketRepo interface {
		BucketWritableRepo
		Foreach(func(k []byte, v []byte) error) error
		Get(name string) (*msg.Bucket, error)
		GetBytes(name string) ([]byte, error)
		List(prefix string, size int) ([]*msg.Bucket, int, error)
	}

	BucketService interface {
		BucketRepo
	}
)
