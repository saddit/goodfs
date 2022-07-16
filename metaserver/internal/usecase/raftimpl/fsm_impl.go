package raftimpl

import (
	"io"
	. "metaserver/internal/usecase"

	"github.com/hashicorp/raft"
)

//TODO 应用元数据写入
type fsm struct {
	IMetadataService
}

func NewFSM(service IMetadataService) raft.FSM {
	return &fsm{service}
}

func (f *fsm) Apply(lg *raft.Log) any {
	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (f *fsm) Restore(snapshot io.ReadCloser) error {
	return nil
}
