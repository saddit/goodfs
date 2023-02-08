package entity

import (
	"apiserver/config"
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

type ObjectStrategy int8

const (
	ECReedSolomon ObjectStrategy = 1 << iota
	MultiReplication
)

type Extra struct {
	Total        int `json:"total"`
	FirstVersion int `json:"firstVersion"`
	LastVersion  int `json:"lastVersion"`
}

type Metadata struct {
	Extra
	Name       string     `json:"name"`
	Bucket     string     `json:"bucket"`
	CreateTime int64      `json:"createTime"`
	UpdateTime int64      `json:"updateTime"`
	Versions   []*Version `json:"versions"`
}

func (m *Metadata) LastVersion() *Version {
	if len(m.Versions) > 0 {
		return m.Versions[len(m.Versions)-1]
	}
	return nil
}

type Version struct {
	Compress      bool           `json:"compress"`
	Hash          string         `json:"hash"`
	StoreStrategy ObjectStrategy `json:"storeStrategy"`
	Sequence      int32          `json:"sequence"`
	Size          int64          `json:"size"`
	Ts            int64          `json:"ts"`
	DataShards    int            `json:"dataShards"`
	ParityShards  int            `json:"parityShards"`
	ShardSize     int            `json:"shardSize"`
	Locate        []string       `json:"locate"`
}

type Bucket struct {
	Versioning     bool           `json:"versioning"`     // Versioning marks bucket can store multi versions of object. if true, VersionRemains will be used
	Readonly       bool           `json:"readonly"`       // Readonly marks objects in bucket only allowed to read
	Compress       bool           `json:"compress"`       // Compress marks objects in bucket should be compressed before store
	StoreStrategy  ObjectStrategy `json:"storeStrategy"`  // StoreStrategy if not zero, it will apply to ever objects under this bucket
	DataShards     int            `json:"dataShards"`     // DataShards used when StoreStrategy is not zero
	ParityShards   int            `json:"parityShards"`   // ParityShards used when StoreStrategy is not zero
	VersionRemains int            `json:"versionRemains"` // VersionRemains is maximum number of remained versions
	CreateTime     int64          `json:"createTime"`     // CreateTime is bucket created time
	UpdateTime     int64          `json:"updateTime"`     // UpdateTime is last updating time
	Name           string         `json:"name"`           // Name is the bucket's name
	Policies       []string       `json:"policies"`       // Policies is the iam polices for this bucket (No support yet)
}

func (b *Bucket) MakeVersion(ver *Version, conf *config.ObjectConfig) {
	if b.Compress {
		ver.Compress = true
	}
	// copy of config
	rsConf, rpConf := conf.ReedSolomon, conf.Replication

	if b.StoreStrategy > 0 {
		ver.StoreStrategy = b.StoreStrategy
		rsConf.DataShards = b.DataShards
		rsConf.ParityShards = b.ParityShards
		rpConf.CopiesCount = b.DataShards
	}

	switch ver.StoreStrategy {
	default:
		fallthrough
	case ECReedSolomon:
		ver.DataShards = rsConf.DataShards
		ver.ParityShards = rsConf.ParityShards
		ver.ShardSize = rsConf.ShardSize(ver.Size)
	case MultiReplication:
		ver.DataShards = rpConf.CopiesCount
		ver.ShardSize = int(ver.Size)
	}
}

func (b *Bucket) MakeConf(conf *config.ObjectConfig) (cfg config.ObjectConfig) {
	// copy of config
	cfg = *conf
	switch b.StoreStrategy {
	case ECReedSolomon:
		cfg.ReedSolomon.DataShards = b.DataShards
		cfg.ReedSolomon.ParityShards = b.ParityShards
	case MultiReplication:
		cfg.Replication.CopiesCount = b.DataShards
	}
	return
}
