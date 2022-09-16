package logic

import (
	"common/hashslot"
	"common/logs"
	"metaserver/internal/usecase/pool"
)

type HashSlot struct{}

func NewHashSlot() HashSlot { return HashSlot{} }

func (h HashSlot) IsKeyOnThisServer(key string) (bool, string) {
	id, err := pool.HashSlot.GetKeyIdentify(key)
	if err != nil {
		logs.Std().Error(err)
		return false, ""
	}
	return id == pool.HttpHostPort, id
}

func (HashSlot) GetSlotsProvider() (hashslot.IEdgeProvider, error) {
	return pool.HashSlot.GetEdgeProvider(false)
}

func (h HashSlot) SaveToEtcd(id string, info *hashslot.SlotInfo) error {
	return pool.HashSlot.Save(id, info)
}

func (HashSlot) RemoveFromEtcd(id string) error {
	return pool.HashSlot.Remove(id)
}
