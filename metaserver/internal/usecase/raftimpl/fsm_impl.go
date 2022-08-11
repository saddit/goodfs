package raftimpl

import (
	"common/graceful"
	"common/logs"
	"common/util"
	"encoding/json"
	"io"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
	"metaserver/internal/usecase/logic"
	"metaserver/internal/usecase/utils"

	"github.com/hashicorp/raft"
	bolt "go.etcd.io/bbolt"
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
		logs.Std().Warn("recieve log type %v", lg.Type)
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
	reader, writer := io.Pipe()
	enc := json.NewEncoder(writer)
	go func() {
		defer graceful.Recover()
		defer writer.Close()
		err := f.View(func(tx *bolt.Tx) error {
			// get root bucket
			b := tx.Bucket([]byte(logic.BucketName))
			// each metadata kv
			b.ForEach(func(k, v []byte) error {
				var data entity.Metadata
				if !utils.DecodeMsgp(&data, v) {
					return ErrDecode
				}
				return enc.Encode(&entity.RaftData{
					Type: entity.LogInsert,
					Dest: entity.DestMetadata,
					Name: data.Name,
					Metadata: &data,
				})
			})
			btx := b.Tx()
			defer btx.Rollback()
			// each metadata-version bunckets
			return btx.ForEach(func(name []byte, b *bolt.Bucket) error {
				// each metadata's versions
				return b.ForEach(func(k, v []byte) error {
					var data entity.Version
					if !utils.DecodeMsgp(&data, v) {
						return ErrDecode
					}
					return enc.Encode(&entity.RaftData{
						Name: string(name),
						Sequence: data.Sequence,
						Type: entity.LogInsert,
						Dest: entity.DestVersion,
						Version: &data,
					})
				})
			})
		})
		util.LogErr(err)
	}()
	return &snapshot{reader}, nil
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
			if data.Type == entity.LogInsert {
				if data.Dest == entity.DestMetadata {
					err = logic.AddMeta(data.Name, data.Metadata)(tx)
				} else {
					err = logic.AddVer(data.Name, data.Version)(tx)
				}
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

type snapshot struct {
	io.Reader
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := io.Copy(sink, s); err != nil {
		logs.Std().Error(err)
		return sink.Cancel()
	}
	return sink.Close()
}

func (s *snapshot) Release() {

}
