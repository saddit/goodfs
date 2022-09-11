package entity

import (
	"time"
)

type VerMode int32

const (
	//VerModeFirst 查询第一个版本
	VerModeFirst VerMode = 1
	//VerModeLast 只查询最后一个版本
	VerModeLast VerMode = 0
	// VerModeNot 不查询任何版本
	VerModeNot VerMode = -1
)

type Metadata struct {
	Name       string     `json:"name"`
	CreateTime time.Time  `json:"createTime"`
	UpdateTime time.Time  `json:"updateTime"`
	Versions   []*Version `json:"versions"`
}

func (m *Metadata) LastVersion() *Version {
	if len(m.Versions) > 0 {
		return m.Versions[len(m.Versions)-1]
	}
	return nil
}

type Version struct {
	Sequence     int32     `json:"sequence"`
	Hash         string    `json:"hash"`
	Size         int64     `json:"size"`
	Ts           time.Time `json:"ts"`
	EcAlgo       int       `json:"ecAlgo"`
	DataShards   int       `json:"dataShards"`
	ParityShards int       `json:"parityShards"`
	ShardSize    int       `json:"shardSize"`
	Locate       []string  `json:"locate"`
}
