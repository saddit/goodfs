package meta

import (
	"time"
)

type MetaData struct {
	Id         string        `bson:"_id,omitempty"`
	Name       string        `bson:"name,omitempty"`
	Tags       []string      `bson:"tags"`
	CreateTime time.Time     `bson:"create_time,omitempty"`
	UpdateTime time.Time     `bson:"update_time,omitempty"`
	Versions   []*MetaVersion `bson:"versions,omitempty"`
}

type MetaVersion struct {
	Hash   string    `bson:"hash,omitempty"`
	Size   int64     `bson:"size"`
	Ts     time.Time `bson:"ts,omitempty"`
	Locate string    `bson:"locate"`
}
