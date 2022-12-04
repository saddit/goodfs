package logic

import (
	"apiserver/internal/usecase/pool"
	"common/cst"
	"common/hashslot"
	"common/response"
	"common/util"
	"context"
	"net/http"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type HashSlot struct {
}

func NewHashSlot() *HashSlot {
	return &HashSlot{}
}

// FindMetaLocOfName find metadata location by hash-slot-algo return master-loc, group id, error
func (HashSlot) FindMetaLocOfName(name string) (string, string, error) {
	if time.Now().Unix()-hashSlotCache.updatedAt < expiredDuration {
		loc, err := hashslot.GetStringIdentify(name, hashSlotCache.provider)
		if err != nil {
			// reset cache if err
			hashSlotCache.reset()
			return "", "", err
		}
		return loc, hashSlotCache.slotIdMap[loc], nil
	}
	slotsMap := make(map[string][]string)
	slotsIdMap := make(map[string]string)
	prefix := cst.EtcdPrefix.FmtHashSlot(pool.Config.Registry.Group, pool.Config.Discovery.MetaServName, "")
	// get slots data from etcd (only master saves into to etcd)
	res, err := pool.Etcd.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return "", "", err
	}
	// wrap slot
	for _, kv := range res.Kvs {
		var info hashslot.SlotInfo
		if err := util.DecodeMsgp(&info, kv.Value); err != nil {
			return "", "", err
		}
		slotsMap[info.Location] = info.Slots
		slotsIdMap[info.Location] = info.GroupID
	}
	slots, err := hashslot.WrapSlots(slotsMap)
	if err != nil {
		return "", "", response.NewError(http.StatusServiceUnavailable, err.Error())
	}
	// update cache sync
	go hashSlotCache.update(slots, slotsIdMap)
	// find location
	loc, err := hashslot.GetStringIdentify(name, slots)
	if err != nil {
		return "", "", err
	}
	return loc, slotsIdMap[loc], nil
}
