package entity

import (
	"common/system"
)

type ServerInfo struct {
	ServerID string       `json:"serverId"`
	HttpAddr string       `json:"httpAddr"`
	SysInfo  *system.Info `json:"sysInfo"`
	IsMaster bool         `json:"isMaster,omitempty"`
}
