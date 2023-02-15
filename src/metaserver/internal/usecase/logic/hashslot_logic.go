package logic

import (
	"common/hashslot"
	"common/logs"
	"metaserver/internal/usecase"
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
	return id == pool.Config.Registry.ServerID, id
}

func (HashSlot) GetSlotsProvider() (hashslot.IEdgeProvider, error) {
	return pool.HashSlot.GetEdgeProvider(false)
}

func (h HashSlot) SaveToEtcd(id string, info *hashslot.SlotInfo) error {
	info.GroupID = id
	info.ServerID = pool.Config.Registry.ServerID
	go func() {
		pool.Config.HashSlot.Slots = info.Slots
		if err := pool.Config.Persist(); err != nil {
			logs.Std().Errorf("persist config err: %s", err)
			return
		}
	}()
	return pool.HashSlot.Save(id, info)
}

func (HashSlot) RemoveFromEtcd(id string) error {
	return pool.HashSlot.Remove(id)
}

func (HashSlot) GetById(storeId string) (*hashslot.SlotInfo, error) {
	data, ok, err := pool.HashSlot.Get(storeId)
	if err != nil {
		return nil, err
	}
	if !ok {
		if storeId == pool.Config.HashSlot.StoreID {
			return &hashslot.SlotInfo{
				GroupID:  pool.Config.HashSlot.StoreID,
				ServerID: pool.Config.Registry.ServerID,
				Slots:    pool.Config.HashSlot.Slots,
			}, nil
		}
		return nil, usecase.ErrNotFound
	}
	return data, nil
}
