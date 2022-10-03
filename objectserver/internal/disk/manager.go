package disk

import "objectserver/internal/db"

//TODO(feat): provide disk info and save or get from db

type Manager struct {
	infoDB *db.DiskInfo
}

func NewManager(infoDB *db.DiskInfo) *Manager {
	return &Manager{infoDB: infoDB}
}
