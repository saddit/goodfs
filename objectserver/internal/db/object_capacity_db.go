package db

import (
	"common/constrant"
	"common/util"
	"context"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/atomic"
	"strings"
)

type ObjectCapacity struct {
	cli        clientv3.KV
	CurrentCap *atomic.Uint64
	CurrentID  string
}

func NewObjectCapacity(c clientv3.KV, id string) *ObjectCapacity {
	return &ObjectCapacity{c, atomic.NewUint64(0), id}
}

func (oc *ObjectCapacity) Save() error {
	key := constrant.EtcdPrefix.FmtObjectCap(oc.CurrentID)
	_, err := oc.cli.Put(context.Background(), key, oc.CurrentCap.String())
	return err
}

func (oc *ObjectCapacity) GetAll() (map[string]uint64, error) {
	resp, err := oc.cli.Get(context.Background(), constrant.EtcdPrefix.ObjectCap, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	res := make(map[string]uint64, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		sp := strings.Split(string(kv.Key), "/")
		key := sp[len(sp)-1]
		res[key] = util.ToUint64(string(kv.Value))
	}
	return res, nil
}

func (oc *ObjectCapacity) Get(s string) (uint64, error) {
	if s == oc.CurrentID {
		return oc.CurrentCap.Load(), nil
	}
	key := constrant.EtcdPrefix.FmtObjectCap(s)
	resp, err := oc.cli.Get(context.Background(), key)
	if err != nil {
		return 0, err
	}
	if len(resp.Kvs) == 0 {
		return 0, errors.New("not exist capacity " + s)
	}
	return util.ToUint64(string(resp.Kvs[0].Value)), nil
}
