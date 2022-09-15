package logic

import (
	"common/hashslot"
	"common/logs"
	"common/util"
	"context"
	"fmt"
	"metaserver/internal/usecase/pool"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type HashSlot struct{}

func NewHashSlot() HashSlot { return HashSlot{} }

func (h HashSlot) IsKeyOnThisServer(key string) (bool, string) {
	slots, err := h.GetSlotsProvider()
	if err != nil {
		logs.Std().Error(err)
		return false, ""
	}
	// get slot's location of this key
	location, err := hashslot.GetStringIdentify(key, slots)
	if err != nil {
		logs.Std().Error(err)
		return false, ""
	}
	return location == pool.HttpHostPort, location
}

func (HashSlot) GetSlotsProvider() (hashslot.IEdgeProvider, error) {
	slotsMap := make(map[string][]string)
	prefix := fmt.Sprint("hashslot/", pool.Config.Registry.Group, "/", pool.Config.Registry.Name, "/")
	// get slots data from etcd (only master saves into to etcd)
	res, err := pool.Etcd.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	// wrap slot
	for _, kv := range res.Kvs {
		var info hashslot.SlotInfo
		if err := util.DecodeMsgp(&info, kv.Value); err != nil {
			return nil, err
		}
		slotsMap[info.Location] = info.Slots
	}
	return hashslot.WrapSlots(slotsMap)
}

func (HashSlot) SaveToEtcd(info *hashslot.SlotInfo) error {
	return nil
}

func (HashSlot) RemoveFromEtcd() error {
	return nil
}

func (HashSlot) AutoMigrate() {

}

func (h HashSlot) OnLeaderChanged(isLeader bool) {
	if err := h.RemoveFromEtcd(); err != nil {
		util.LogErr(err)
		return
	}
	if isLeader {
		var info hashslot.SlotInfo
		info.Location = pool.HttpHostPort
		info.Slots = pool.Config.HashSlot
		//TODO get http host:port of peers
		h.SaveToEtcd(&info)
	}
}
