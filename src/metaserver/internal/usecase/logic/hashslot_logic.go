package logic

import (
	"common/hashslot"
	"common/logs"
	"common/util/crypto"
	"metaserver/internal/usecase/pool"
	"strings"
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
	info.GroupID = id
	info.Location = pool.HttpHostPort
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
		return &hashslot.SlotInfo{
			GroupID:  pool.Config.HashSlot.StoreID,
			ServerID: pool.Config.Registry.ServerID,
			Location: pool.HttpHostPort,
			Checksum: crypto.MD5([]byte(strings.Join(pool.Config.HashSlot.Slots, ","))),
			Slots:    pool.Config.HashSlot.Slots,
		}, nil
	}
	return data, nil
}
