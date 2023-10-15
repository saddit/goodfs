package msg

import (
	"common/util"
	"fmt"
)

//go:generate msgp -tests=false #msg

type Extra struct {
	Total        int `json:"total" msg:"total"`
	FirstVersion int `json:"firstVersion" msg:"first_version"`
	LastVersion  int `json:"lastVersion" msg:"last_version"`
}

type Metadata struct {
	*Extra     `msg:",inline"`
	Name       string `json:"name" msg:"name" binding:"required"`
	Bucket     string `json:"bucket" msg:"bucket" binding:"required"`
	CreateTime int64  `json:"createTime" msg:"create_time"`
	UpdateTime int64  `json:"updateTime" msg:"update_time"`
}

func (z *Metadata) ID() string {
	return fmt.Sprintf(z.Bucket, "/", z.Name)
}

type Version struct {
	Compress      bool     `json:"compress" msg:"compress"`
	StoreStrategy int8     `json:"storeStrategy" msg:"store_strategy" binding:"required"`
	DataShards    int32    `json:"dataShards" msg:"data_shards" binding:"required"`
	ParityShards  int32    `json:"parityShards" msg:"parity_shards"`
	ShardSize     int64    `json:"shardSize" msg:"shard_size" binding:"required"`
	Size          int64    `json:"size" msg:"size" binding:"required"`
	Ts            int64    `json:"ts" msg:"ts"`
	Sequence      uint64   `json:"sequence" msg:"sequence"` // Sequence version number auto generated on saving
	Hash          string   `json:"hash" msg:"hash" binding:"required"`
	UniqueId      string   `json:"uniqueId" msg:"uniqueId"`
	Locate        []string `json:"locate" msg:"locate" binding:"min=1"`
}

func (z *Version) ID() string {
	return util.UIntString(z.Sequence)
}

type Bucket struct {
	Versioning     bool     `json:"versioning" msg:"versioning"`          // Versioning marks bucket can store multi versions of object. if true, VersionRemains will be used
	Readonly       bool     `json:"readonly" msg:"readonly"`              // Readonly marks objects in bucket only allowed to read
	Compress       bool     `json:"compress" msg:"compress"`              // Compress marks objects in bucket should be compressed before store
	StoreStrategy  int8     `json:"storeStrategy" msg:"store_strategy"`   // StoreStrategy if not zero, it will apply to ever objects under this bucket
	DataShards     int32    `json:"dataShards" msg:"data_shards"`         // DataShards used when StoreStrategy is not zero
	ParityShards   int32    `json:"parityShards" msg:"parity_shards"`     // ParityShards used when StoreStrategy is not zero
	VersionRemains int32    `json:"versionRemains" msg:"version_remains"` // VersionRemains is maximum number of remained versions
	CreateTime     int64    `json:"createTime" msg:"create_time"`         // CreateTime is bucket created time
	UpdateTime     int64    `json:"updateTime" msg:"update_time"`         // UpdateTime is last updating time
	Name           string   `json:"name" msg:"name"`                      // Name is the bucket's name
	Policies       []string `json:"policies" msg:"policies"`              // Policies is the iam polices for this bucket (No support yet)
}

func (z *Bucket) ID() string {
	return z.Name
}
