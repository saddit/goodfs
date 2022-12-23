package entity

//go:generate msgp -tests=false

type Metadata struct {
	Name       string `json:"name" msg:"name" binding:"required"`
	CreateTime int64  `json:"createTime" msg:"create_time"`
	UpdateTime int64  `json:"updateTime" msg:"update_time"`
}

type Version struct {
	Sequence      uint64   `json:"sequence" msg:"sequence"` //Sequence version number auto generated on saving
	Hash          string   `json:"hash" msg:"hash" binding:"required"`
	Size          int64    `json:"size" msg:"size" binding:"required"`
	Ts            int64    `json:"ts" msg:"ts"`
	StoreStrategy int8     `json:"storeStrategy" msg:"store_strategy" binding:"required"`
	DataShards    int32    `json:"dataShards" msg:"data_shards" binding:"required"`
	ParityShards  int32    `json:"parityShards" msg:"parity_shards" binding:"required"`
	ShardSize     int64    `json:"shardSize" msg:"shard_size" binding:"required"`
	Locate        []string `json:"locate" msg:"locate" binding:"min=1"`
}
