package entity

import (
	"time"
)

type VerMode int32

const (
	//VerModeALL 查询全部版本
	VerModeALL VerMode = -128
	//VerModeLast 只查询最后一个版本
	VerModeLast VerMode = -2
	// VerModeNot 不查询任何版本
	VerModeNot VerMode = -1
)

type MetaData struct {
	Id         string     `bson:"_id,omitempty" json:"id"`
	Name       string     `bson:"name,omitempty" json:"name"`
	Tags       []string   `bson:"tags" json:"tags"`
	CreateTime time.Time  `bson:"create_time,omitempty" json:"createTime"`
	UpdateTime time.Time  `bson:"update_time,omitempty" json:"updateTime"`
	Versions   []*Version `bson:"versions,omitempty" json:"versions"`
}

type Version struct {
	Hash   string    `bson:"hash,omitempty" json:"hash"`
	Size   int64     `bson:"size" json:"size"`
	Ts     time.Time `bson:"ts,omitempty" json:"ts"`
	Locate []string  `bson:"locate" json:"locate"`
}
