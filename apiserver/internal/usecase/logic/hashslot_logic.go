package logic

import (
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/selector"
	"common/constrant"
	"common/hashslot"
	"common/util"
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type HashSlot struct {
}

func NewHashSlot() *HashSlot {
	return &HashSlot{}
}

// FindMetaLocOfName find metadata location by hash-slot-algo（load balance）
func (HashSlot) FindMetaLocOfName(name string) (string, error) {
	slotsMap := make(map[string][]string)
	prefix := constrant.EtcdPrefix.FmtHashSlot(pool.Config.Registry.Group, pool.Config.Discovery.MetaServName, "")
	// get slots data from etcd (only master saves into to etcd)
	res, err := pool.Etcd.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	// wrap slot
	for _, kv := range res.Kvs {
		var info hashslot.SlotInfo
		if err := util.DecodeMsgp(&info, kv.Value); err != nil {
			return "", err
		}
		// load balance with slaves
		lb := selector.NewIPSelector(pool.Balancer, info.Peers)
		slotsMap[lb.Select()] = info.Slots
	}
	slots, err := hashslot.WrapSlots(slotsMap)
	if err != nil {
		return "", err
	}
	loc, err := hashslot.GetStringIdentify(name, slots)
	if err != nil {
		return "", err
	}
	return loc, nil
}
