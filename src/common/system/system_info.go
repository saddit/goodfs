package system

import (
	"common/system/disk"
	"common/system/mem"
)

//go:generate msgp -tests=false

type Info struct {
	DiskInfo  disk.Info  `json:"diskInfo" msg:",inline"`
	MemStatus mem.Status `json:"memStatus" msg:",inline"`
}

func NewInfo(diskPath string) (*Info, error) {
	d, err := disk.GetInfo(diskPath)
	if err != nil {
		return nil, err
	}
	return &Info{d, mem.MemStat()}, nil
}
