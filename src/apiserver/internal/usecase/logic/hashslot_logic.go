package logic

import (
	"apiserver/internal/usecase/pool"
	"common/cst"
	"common/hashslot"
	"common/response"
	"common/util"
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
)

type HashSlot struct {
}

func NewHashSlot() *HashSlot {
	return &HashSlot{}
}

// KeySlotLocation find metadata location by hash-slot-algo return master server id, error
func (HashSlot) KeySlotLocation(name string) (string, error) {
	slotsMap := make(map[string][]string)
	prefix := cst.EtcdPrefix.FmtHashSlot(pool.Config.Registry.Group, pool.Config.Discovery.MetaServName, "")
	// get slots data from etcd (only master saves into to etcd)
	res, err := pool.Etcd.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return "", err
	}
	// wrap slot
	for _, kv := range res.Kvs {
		var info hashslot.SlotInfo
		if err = util.DecodeMsgp(&info, kv.Value); err != nil {
			return "", err
		}
		slotsMap[info.ServerID] = info.Slots
	}
	slots, err := hashslot.WrapSlots(slotsMap)
	if err != nil {
		return "", response.NewError(http.StatusServiceUnavailable, err.Error())
	}
	sid, err := hashslot.GetStringIdentify(name, slots)
	if err != nil {
		return "", err
	}
	return sid, nil
}
