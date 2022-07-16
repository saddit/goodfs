package entity

//go:generate msgp

type Metadata struct {
	Name       string `json:"name" msg:"name"`
	CreateTime int64  `json:"createTime" msg:"create_time"`
	UpdateTime int64  `json:"updateTime" msg:"update_time"`
}

type Version struct {
	Sequence     uint64   `json:"sequence" msg:"sequence"` //Sequence version number auto generated on saving
	Hash         string   `json:"hash" msg:"hash"`
	Size         int64    `json:"size" msg:"size"`
	Ts           int64    `json:"ts" msg:"ts"`
	EcAlgo       int8     `json:"ecAlgo" msg:"ec_algo"`
	DataShards   int32    `json:"dataShards" msg:"data_shards"`
	ParityShards int32    `json:"parityShards" msg:"parity_shards"`
	ShardSize    int64    `json:"shardSize" msg:"shard_size"`
	Locate       []string `json:"locate" msg:"locate"`
}
