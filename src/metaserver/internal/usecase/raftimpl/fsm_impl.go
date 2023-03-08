package raftimpl

import (
	"common/logs"
	"common/response"
	"common/util"
	"compress/gzip"
	"fmt"
	"github.com/hashicorp/raft"
	"io"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
)

var (
	fsmLog = logs.New("raft-fsm")
)

type FSMImpl struct {
	metaRepo    IMetadataRepo
	metaBatch   IBatchMetaRepo
	bucketRepo  BucketRepo
	bucketBatch BatchBucketRepo
	snapshot    SnapshotManager
}

func NewFSM(m IMetadataRepo, mb IBatchMetaRepo, b BucketRepo, bb BatchBucketRepo, sm SnapshotManager) raft.FSM {
	return &FSMImpl{
		metaRepo:    m,
		metaBatch:   mb,
		bucketRepo:  b,
		bucketBatch: bb,
		snapshot:    sm,
	}
}

func (f *FSMImpl) applyBucket(data *entity.RaftData) *FSMResponse {
	repo := util.IfElse[BucketWritableRepo](data.Batch, f.bucketBatch, f.bucketRepo)
	switch data.Type {
	case entity.LogInsert:
		return FSMResult(repo.Create(data.Bucket))
	case entity.LogRemove:
		return FSMResult(repo.Remove(data.Name))
	case entity.LogUpdate:
		return FSMResult(repo.Update(data.Bucket))
	default:
		return FSMResult(ErrUnknownRaftLog)
	}
}

func (f *FSMImpl) applyMetadata(data *entity.RaftData) *FSMResponse {
	repo := util.IfElse[WritableRepo](data.Batch, f.metaBatch, f.metaRepo)
	switch data.Type {
	case entity.LogInsert:
		return FSMResult(repo.AddMetadata(data.Name, data.Metadata))
	case entity.LogRemove:
		return FSMResult(repo.RemoveMetadata(data.Name))
	case entity.LogUpdate:
		return FSMResult(repo.UpdateMetadata(data.Name, data.Metadata))
	default:
		return FSMResult(ErrUnknownRaftLog)
	}
}

func (f *FSMImpl) applyVersion(data *entity.RaftData) *FSMResponse {
	repo := util.IfElse[WritableRepo](data.Batch, f.metaBatch, f.metaRepo)
	switch data.Type {
	case entity.LogMigrate:
		resp := FSMResult(repo.AddVersionWithSequence(data.Name, data.Version))
		return resp
	case entity.LogInsert:
		resp := FSMResult(repo.AddVersion(data.Name, data.Version))
		resp.Data = data.Version.Sequence
		return resp
	case entity.LogRemove:
		return FSMResult(repo.RemoveVersion(data.Name, data.Sequence))
	case entity.LogUpdate:
		data.Version.Sequence = data.Sequence
		return FSMResult(repo.UpdateVersion(data.Name, data.Version))
	default:
		return FSMResult(ErrUnknownRaftLog)
	}
}

func (f *FSMImpl) applyVersionAll(data *entity.RaftData) *FSMResponse {
	repo := util.IfElse[WritableRepo](data.Batch, f.metaBatch, f.metaRepo)
	switch data.Type {
	case entity.LogRemove:
		return FSMResult(repo.RemoveAllVersion(data.Name))
	default:
		return FSMResult(ErrUnknownRaftLog)
	}
}

func (f *FSMImpl) Apply(lg *raft.Log) (r any) {
	if lg == nil || len(lg.Data) == 0 {
		return FSMResult(ErrNilData)
	}
	if lg.Type != raft.LogCommand {
		return fmt.Errorf("drop recieved fsmLog type %v", lg.Type)
	}

	lst, err := f.snapshot.LastAppliedIndex()
	if err != nil {
		return err
	}
	fsmLog.Debugf("apply log index %d and recorded index is %d", lg.Index, lst)
	if lst >= lg.Index {
		return nil
	}

	defer func() {
		util.LogErrWithPre("fsm record apply index", f.snapshot.ApplyIndex(lg.Index))
	}()

	var data entity.RaftData
	if err := util.DecodeMsgp(&data, lg.Data); err != nil {
		return err
	}

	switch data.Dest {
	case entity.DestMetadata:
		return f.applyMetadata(&data)
	case entity.DestVersion:
		return f.applyVersion(&data)
	case entity.DestVersionAll:
		return f.applyVersionAll(&data)
	case entity.DestBucket:
		return f.applyBucket(&data)
	}
	return ErrUnknownRaftLog
}

func (f *FSMImpl) ApplyBatch(lgs []*raft.Log) []any {
	var data entity.RaftData
	res := make([]any, len(lgs))

	lastAppliedIndex, err := f.snapshot.LastAppliedIndex()
	if err != nil {
		for i := range res {
			res[i] = err
		}
		return res
	}
	var maxIndex uint64
	for i, lg := range lgs {
		if lg == nil || len(lg.Data) == 0 {
			res[i] = FSMResult(ErrNilData)
			continue
		}
		if lg.Type != raft.LogCommand {
			res[i] = fmt.Errorf("drop recieved fsmLog type %v", lg.Type)
			continue
		}
		fsmLog.Debugf("apply log index %d and recorded index is %d", lg.Index, lastAppliedIndex)
		if lastAppliedIndex >= lg.Index {
			continue
		}
		if err = util.DecodeMsgp(&data, lg.Data); err != nil {
			res[i] = err
			continue
		}
		data.Batch = true
		switch data.Dest {
		case entity.DestMetadata:
			res[i] = f.applyMetadata(&data)
		case entity.DestVersion:
			res[i] = f.applyVersion(&data)
		case entity.DestVersionAll:
			res[i] = f.applyVersionAll(&data)
		case entity.DestBucket:
			res[i] = f.applyBucket(&data)
		default:
			res[i] = ErrUnknownRaftLog
		}

		if lg.Index > maxIndex {
			maxIndex = lg.Index
		}
	}
	//NOTICE: metaBatch Sync and bucketBatch Sync it's same for now.
	if err = f.metaBatch.Sync(); err != nil {
		err = fmt.Errorf("sync fail: %s", err)
		for i := range res {
			if res[i] == nil {
				res[i] = err
			}
		}
	} else if maxIndex > lastAppliedIndex {
		util.LogErrWithPre("fsm record apply index", f.snapshot.ApplyIndex(maxIndex))
	}
	return res
}

func (f *FSMImpl) Snapshot() (raft.FSMSnapshot, error) {
	snap, err := f.snapshot.Snapshot()
	if err != nil {
		return nil, err
	}
	return &snapshot{snap}, nil
}

func (f *FSMImpl) Restore(snapshot io.ReadCloser) (err error) {
	defer snapshot.Close()
	gzipRd, err := gzip.NewReader(snapshot)
	if err != nil {
		return err
	}
	defer gzipRd.Close()
	return f.snapshot.Restore(gzipRd)
}

type snapshot struct {
	SnapshotTx
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	gzipWt := gzip.NewWriter(sink)
	if _, err := s.WriteTo(gzipWt); err != nil {
		fsmLog.Error(err)
		return sink.Cancel()
	}
	util.LogErr(gzipWt.Close())
	return sink.Close()
}

func (s *snapshot) Release() {
	util.LogErr(s.Rollback())
}

type FSMResponse struct {
	err  *response.Err
	Data any
}

func FSMResult(err error) *FSMResponse {
	switch err := err.(type) {
	case *response.Err:
		return &FSMResponse{err, nil}
	case response.Err:
		return &FSMResponse{&err, nil}
	case nil:
		return &FSMResponse{&response.Err{Status: 200}, nil}
	default:
		return &FSMResponse{&response.Err{Status: 500, Message: err.Error()}, nil}
	}
}

func (r *FSMResponse) Ok() bool {
	return r.err.Status/100 == 2
}

func (r *FSMResponse) ToError() error {
	if r.Ok() {
		return nil
	}
	return r.err
}
