package meta

import (
	"time"
)

type MetaDataV2 struct {
	Id         string           `bson:"_id,omitempty" json:"id"`
	Name       string           `bson:"name,omitempty" json:"name"`
	Tags       []string         `bson:"tags" json:"tags"`
	CreateTime time.Time        `bson:"create_time,omitempty" json:"createTime"`
	UpdateTime time.Time        `bson:"update_time,omitempty" json:"updateTime"`
	Versions   []*MetaVersionV2 `bson:"versions,omitempty" json:"versions"`
}

type MetaVersionV2 struct {
	Hash   string    `bson:"hash,omitempty" json:"hash"`
	Size   int64     `bson:"size" json:"size"`
	Ts     time.Time `bson:"ts,omitempty" json:"ts"`
	Locate []string  `bson:"locate" json:"locate"`
}
