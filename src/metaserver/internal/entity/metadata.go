package entity

//go:generate msgp -tests=false

type Metadata struct {
	Name       string `json:"name" msg:"name" binding:"required"`
	Bucket     string `json:"bucket" msg:"bucket"`
	CreateTime int64  `json:"createTime" msg:"create_time"`
	UpdateTime int64  `json:"updateTime" msg:"update_time"`
}

type Version struct {
	StoreStrategy int8     `json:"storeStrategy" msg:"store_strategy" binding:"required"`
	DataShards    int32    `json:"dataShards" msg:"data_shards" binding:"required"`
	ParityShards  int32    `json:"parityShards" msg:"parity_shards"`
	ShardSize     int64    `json:"shardSize" msg:"shard_size" binding:"required"`
	Size          int64    `json:"size" msg:"size" binding:"required"`
	Ts            int64    `json:"ts" msg:"ts"`
	Sequence      uint64   `json:"sequence" msg:"sequence"` // Sequence version number auto generated on saving
	Hash          string   `json:"hash" msg:"hash" binding:"required"`
	Locate        []string `json:"locate" msg:"locate" binding:"min=1"`
}

type Bucket struct {
	Versioning     bool     `json:"versioning" msg:"versioning"`                     // Versioning marks bucket can store multi versions of object. if true, VersionRemains will be used
	Readonly       bool     `json:"readonly" msg:"readonly"`                         // Readonly marks objects in bucket only allowed to read
	StoreStrategy  int8     `json:"storeStrategy" msg:"store_strategy"`              // StoreStrategy if not zero, it will apply to ever objects under this bucket
	DataShards     int32    `json:"dataShards" msg:"data_shards" binding:"required"` // DataShards used when StoreStrategy is not zero
	ParityShards   int32    `json:"parityShards" msg:"parity_shards"`                // ParityShards used when StoreStrategy is not zero
	VersionRemains int32    `json:"versionRemains" msg:"version_remains"`            // VersionRemains is maximum number of remained versions
	CreateTime     int64    `json:"createTime" msg:"create_time"`                    // CreateTime is bucket created time
	UpdateTime     int64    `json:"updateTime" msg:"update_time"`                    // UpdateTime is last updating time
	Name           string   `json:"name" msg:"name"`                                 // Name is the bucket's name
	Policies       []string `json:"policies" msg:"policies"`                         // Policies is the iam polices for this bucket (No support yet)
}
