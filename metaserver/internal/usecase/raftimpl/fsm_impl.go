package raftimpl

import (
	"encoding/json"
	"io"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
	"metaserver/internal/usecase/logic"
	bolt "go.etcd.io/bbolt"
	"github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"
)

type fsm struct {
	ITransaction
}

func NewFSM(tx ITransaction) raft.FSM {
	return &fsm{tx}
}

func (f *fsm) applyMetadata(data *entity.RaftData) error {
	switch data.Type {
	case entity.LogInsert:
		return f.Update(logic.AddMeta(data.Name, data.Metadata)) 
	case entity.LogRemove:
		return f.Update(logic.RemoveMeta(data.Name))
	case entity.LogUpdate:
		return f.Update(logic.UpdateMeta(data.Name, data.Metadata))
	default:
		return ErrNotFound
	}
}

func (f *fsm) applyVersion(data *entity.RaftData) error {
	switch data.Type {
	case entity.LogInsert:
		return f.Update(logic.AddVer(data.Name, data.Version))
	case entity.LogRemove:
		return f.Update(logic.RemoveVer(data.Name, int(data.Sequence)))
	case entity.LogUpdate:
		data.Version.Sequence = data.Sequence
		return f.Update(logic.UpdateVer(data.Name, data.Version))
	default:
		return ErrNotFound
	}
}

func (f *fsm) Apply(lg *raft.Log) any {
	if lg.Type != raft.LogCommand {
		logrus.Warn("recieve log type %v", lg.Type)
		return nil
	}
	var data entity.RaftData
	if err := json.Unmarshal(lg.Data, &data); err != nil {
		return err
	}
	if data.Dest == entity.DestMetadata {
		return f.applyMetadata(&data)
	} else if data.Dest == entity.DestVersion {
		return f.applyVersion(&data)
	}
	return ErrNotFound
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return &snapshot{}, nil
}


func (f *fsm) Restore(snapshot io.ReadCloser) error {
	dc := json.NewDecoder(snapshot)
	return f.Update(func(tx *bolt.Tx) error {
		var err error
		for dc.More() {
			// decode data
			var data entity.RaftData
			if err = dc.Decode(&data); err != nil {
				return err
			}
			// judge data type and dest then do logic
			switch data.Type {
			case entity.LogInsert:
				if data.Dest == entity.DestMetadata {
					err = logic.AddMeta(data.Name, data.Metadata)(tx)
				} else {
					err = logic.AddVer(data.Name, data.Version)(tx)
				}
			case entity.LogUpdate:
				if data.Dest == entity.DestMetadata {
					err = logic.UpdateMeta(data.Name, data.Metadata)(tx)
				} else {
					err = logic.UpdateVer(data.Name, data.Version)(tx)
				}
			case entity.LogRemove:
				if data.Dest == entity.DestMetadata {
					err = logic.RemoveMeta(data.Name)(tx)
				} else {
					err = logic.RemoveVer(data.Name, int(data.Sequence))(tx)
				}
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
}

type snapshot struct {
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	return nil
}

func (s *snapshot) Release() {

}
