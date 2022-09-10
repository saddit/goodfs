package logic

import (
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/pool"
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

func (HashSlot) FindLocOfName(name string, locations []string) (string, error) {
	slotsMap := make(map[string][]string)
	for _, loc := range locations {
		resp, err := pool.Etcd.Get(context.Background(), fmt.Sprint("metaserver_hashslot/", loc), clientv3.WithFirstKey()...)
		if err != nil {
			return "", err
		}
		if len(resp.Kvs) == 0 {
			return "", usecase.ErrServiceUnavailable
		}
		slotsMap[loc] = strings.Split(string(resp.Kvs[0].Value), ",")
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
