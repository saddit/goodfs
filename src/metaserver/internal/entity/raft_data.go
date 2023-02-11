package entity

import "common/proto/msg"

//go:generate msgp -tests=false #metaserver/entity

type LogType int8
type Dest int8

const (
	LogInsert LogType = 1 << iota
	LogRemove
	LogUpdate
	LogMigrate
)

const (
	DestVersion Dest = 1 << iota
	DestVersionAll
	DestMetadata
	DestBucket
)

type RaftData struct {
	Type     LogType       `msg:"type" json:"type"`
	Dest     Dest          `msg:"dest" json:"dest"`
	Name     string        `msg:"name" json:"name"`
	Sequence uint64        `msg:"sequence" json:"sequence,omitempty"`
	Version  *msg.Version  `msg:"version" json:"version,omitempty"`
	Metadata *msg.Metadata `msg:"metadata" json:"metadata,omitempty"`
	Bucket   *msg.Bucket   `msg:"bucket" json:"bucket,omitempty"`
	Batch    bool          `msg:"-" json:"-"`
}
