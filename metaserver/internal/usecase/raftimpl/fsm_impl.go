package raftimpl

import (
	"common/logs"
	"common/util"
	"io"
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
	"metaserver/internal/usecase/logic"
	"metaserver/internal/usecase/utils"
	"os"

	"github.com/hashicorp/raft"
	bolt "go.etcd.io/bbolt"
)

var (
	log = logs.New("raft-fsm")
)

type fsm struct {
	db *bolt.DB
}

func NewFSM(tx *bolt.DB) raft.FSM {
	return &fsm{tx}
}

func (f *fsm) applyMetadata(data *entity.RaftData) error {
	switch data.Type {
	case entity.LogInsert:
		return f.db.Update(logic.AddMeta(data.Name, data.Metadata))
	case entity.LogRemove:
		return f.db.Update(logic.RemoveMeta(data.Name))
	case entity.LogUpdate:
		return f.db.Update(logic.UpdateMeta(data.Name, data.Metadata))
	default:
		return ErrNotFound
	}
}

func (f *fsm) applyVersion(data *entity.RaftData) error {
	switch data.Type {
	case entity.LogInsert:
		return f.db.Update(logic.AddVer(data.Name, data.Version))
	case entity.LogRemove:
		return f.db.Update(logic.RemoveVer(data.Name, int(data.Sequence)))
	case entity.LogUpdate:
		data.Version.Sequence = data.Sequence
		return f.db.Update(logic.UpdateVer(data.Name, data.Version))
	default:
		return ErrNotFound
	}
}

func (f *fsm) Apply(lg *raft.Log) any {
	if lg.Type != raft.LogCommand {
		log.Warn("recieve log type %v", lg.Type)
		return nil
	}
	var data entity.RaftData
	if ok := utils.DecodeMsgp(&data, lg.Data); !ok {
		return ErrDecode
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
	go func() {
		defer writer.Close()
		tx, err := f.db.Begin(false)
		if err != nil {
			log.Error("get snapshot from fsm err: %v", err)
			return
		}
		defer tx.Rollback()
		n, err := tx.WriteTo(writer)
		if err != nil {
			log.Error("write to snapshot error: %v, written %d", err, n)
			return
		}
	}()
	return &snapshot{reader}, nil
}

func (f *fsm) Restore(snapshot io.ReadCloser) (err error) {
	defer func() {
		if err != nil {
			//TODO re-open db
		}
	}()
	// FIXME close directly may cause panic
	if err = f.db.Close(); err != nil {
		log.Error("restore fail on close db: %v", err)
		return err
	}
	dbPath := f.db.Path()
	if err = os.Rename(dbPath, dbPath+".bak"); err != nil {
		log.Error("restore fail on rename old db file: %v", err)
		return err	
	}
	newFile, err := os.OpenFile(dbPath, os.O_WRONLY | os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Error("restore fail on open new file: %v", err)
		return err
	}
	defer newFile.Close()
	n, err := io.Copy(newFile, snapshot)
	if err != nil {
		log.Error("restore fail on copy data to new file: %v, written %d", err, n)
		return err
	}
	// TODO open new db
	return
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
