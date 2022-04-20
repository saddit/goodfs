package model

const (
	SyncInsert = "insert"
)

const TempKeyPrefix = "TempInfo_"

type TempPostReq struct {
	Name string `uri:"name" binding:"required"`
	Size int64  `header:"Size" binding:"required"`
}

type TempInfo struct {
	Name string
	Id   string
	Size int64
}
