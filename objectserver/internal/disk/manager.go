package disk

import "objectserver/internal/db"

//TODO(feat): provide disk info and save or get from db

type Manager struct {
	infoDB *db.ObjectCapacity
}

func NewManager(infoDB *db.ObjectCapacity) *Manager {
	return &Manager{infoDB: infoDB}
}
