package entity

import (
	"common/datasize"
	"common/system"
)

type ServerInfo struct {
	ServerID string       `json:"serverId"`
	HttpAddr string       `json:"httpAddr"`
	SysInfo  *system.Info `json:"sysInfo"`
	IsMaster bool         `json:"isMaster,omitempty"`
}

type EtcdStatus struct {
	DBSize       datasize.DataSize `json:"dbSize"`
	DBSizeInUse  datasize.DataSize `json:"dbSizeInUse"`
	AlarmMessage []string          `json:"alarmMessage"`
	Endpoint     string            `json:"endpoint"`
	IsLearner    bool              `json:"isLearner"`
}
