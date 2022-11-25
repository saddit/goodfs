package raftimpl

import (
	"common/logs"
	"common/response"
	"common/util"
	"io"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"

	"github.com/hashicorp/raft"
)

var (
	log = logs.New("raft-fsm")
)

type fsm struct {
	repo IMetadataRepo
}

func NewFSM(repo IMetadataRepo) raft.FSM {
	return &fsm{repo}
}

func (f *fsm) applyMetadata(data *entity.RaftData) *response.RaftFsmResp {
	switch data.Type {
	case entity.LogInsert:
		return response.NewRaftFsmResp(f.repo.AddMetadata(data.Metadata))
	case entity.LogRemove:
		return response.NewRaftFsmResp(f.repo.RemoveMetadata(data.Name))
	case entity.LogUpdate:
		return response.NewRaftFsmResp(f.repo.UpdateMetadata(data.Name, data.Metadata))
	default:
		return response.NewRaftFsmResp(ErrNotFound)
	}
}

func (f *fsm) applyVersion(data *entity.RaftData) *response.RaftFsmResp {
	switch data.Type {
	case entity.LogMigrate:
		resp := response.NewRaftFsmResp(f.repo.AddVersionWithSequnce(data.Name, data.Version))
		return resp
	case entity.LogInsert:
		resp := response.NewRaftFsmResp(f.repo.AddVersion(data.Name, data.Version))
		resp.Data = data.Version.Sequence
		return resp
	case entity.LogRemove:
		return response.NewRaftFsmResp(f.repo.RemoveVersion(data.Name, data.Sequence))
	case entity.LogUpdate:
		data.Version.Sequence = data.Sequence
		return response.NewRaftFsmResp(f.repo.UpdateVersion(data.Name, data.Version))
	default:
		return response.NewRaftFsmResp(ErrNotFound)
	}
}

func (f *fsm) applyVersionAll(data *entity.RaftData) *response.RaftFsmResp {
	switch data.Type {
	case entity.LogRemove:
		return response.NewRaftFsmResp(f.repo.RemoveAllVersion(data.Name))
	default:
		return response.NewRaftFsmResp(ErrNotFound)
	}
}

func (f *fsm) Apply(lg *raft.Log) any {
	if lg == nil || len(lg.Data) == 0 {
		return response.NewRaftFsmResp(ErrNilData)
	}
	if lg.Type != raft.LogCommand {
		log.Warn("recieve log type %v", lg.Type)
		return nil
	}
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
	}
	return ErrNotFound
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	reader, _, err := f.repo.ReadDB()
	if err != nil {
		return nil, err
	}
	return &snapshot{reader}, nil
}

func (f *fsm) Restore(snapshot io.ReadCloser) (err error) {
	return f.repo.ReplaceDB(snapshot)
}

type snapshot struct {
	io.ReadCloser
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := io.Copy(sink, s); err != nil {
		log.Error(err)
		return sink.Cancel()
	}
	return sink.Close()
}

func (s *snapshot) Release() {
	util.LogErr(s.Close())
}
