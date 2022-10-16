package entity

//go:generate msgp -tests=false

type LogType int8
type Dest int8

const (
	LogInsert LogType = 1 << iota
	LogRemove
	LogUpdate
)

const (
	DestVersion Dest = 1 << iota
	DestVersionAll
	DestMetadata
)

type RaftData struct {
	Type     LogType   `msg:"type" json:"type"`
	Dest     Dest      `msg:"dest" json:"dest"`
	Name     string    `msg:"name" json:"name"`
	Sequence uint64    `msg:"sequnce" json:"sequence,omitempty"`
	Version  *Version  `msg:"version" json:"version,omitempty"`
	Metadata *Metadata `msg:"metadata" json:"metadata,omitempty"`
}
