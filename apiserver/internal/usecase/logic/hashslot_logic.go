package logic

import (
	"apiserver/internal/usecase/pool"
	"common/collection/set"
	"common/hashslot"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strings"
)

type HashSlot struct {
}

func NewHashSlot() *HashSlot {
	return &HashSlot{}
}

func (HashSlot) FindMetaLocOfName(name string, locations []string) (string, error) {
	validLocs := set.OfString(locations)
	slotsMap := make(map[string][]string)
	prefix := fmt.Sprint(pool.Config.Registry.Group, "/", "hash_slot_", pool.Config.Discovery.MetaServName, "/")
	// get slots data from etcd
	res, err := pool.Etcd.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	// wrap slot
	for _, kv := range res.Kvs {
		keySplit := strings.Split(string(kv.Key), "/")
		identify := keySplit[len(keySplit)-1]
		if validLocs.Contains(identify) {
			slots := strings.Split(string(kv.Value), ",")
			slotsMap[identify] = slots
		}
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
