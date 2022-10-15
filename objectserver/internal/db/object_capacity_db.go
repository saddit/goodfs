package db

import (
	"common/constrant"
	"common/disk"
	"common/graceful"
	"common/logs"
	"common/util"
	"context"
	"errors"
	"objectserver/config"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/atomic"
)

type ObjectCapacity struct {
	cli         clientv3.KV
	CurrentCap  *atomic.Uint64
	CurrentID   string
	groupName   string
	serviceName string
}

func NewObjectCapacity(c clientv3.KV, cfg *config.Config) *ObjectCapacity {
	return &ObjectCapacity{
		c,
		atomic.NewUint64(0),
		cfg.Registry.ServerID,
		cfg.Registry.Group,
		cfg.Registry.Name,
	}
}

func (oc *ObjectCapacity) StartAutoSave(interval time.Duration) func() {
	ctx, cancel := context.WithCancel(context.Background())
	tk := time.NewTicker(interval)
	go func() {
		defer graceful.Recover()
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				util.LogErrWithPre("auto save object-cap", oc.Save())
			case <-ctx.Done():
				logs.Std().Info("stop auto save object-cap and disk-info")
				return
			}
		}
	}()
	return cancel
}

func (oc *ObjectCapacity) Save() error {
	var err error
	var wg sync.WaitGroup
	// save object-cap
	wg.Add(1)
	go func() {
		defer wg.Done()
		keyCap := constrant.EtcdPrefix.FmtObjectCap(oc.groupName, oc.serviceName, oc.CurrentID)
		_, err = oc.cli.Put(context.Background(), keyCap, oc.CurrentCap.String())
	}()
	// save disk-info
	wg.Add(1)
	go func() {
		defer wg.Done()
		var info disk.Info
		var bt []byte
		info, err = oc.CurDiskInfo()
		if err != nil {
			return
		}
		bt, err = util.EncodeMsgp(&info)
		if err != nil {
			return
		}
		keyDisk := constrant.EtcdPrefix.FmtDiskInfo(oc.groupName, oc.serviceName, oc.CurrentID)
		_, err = oc.cli.Put(context.Background(), keyDisk, string(bt))
	}()
	wg.Wait()
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
	key := constrant.EtcdPrefix.FmtObjectCap(oc.groupName, oc.serviceName, oc.CurrentID)
	resp, err := oc.cli.Get(context.Background(), key)
	if err != nil {
		return 0, err
	}
	if len(resp.Kvs) == 0 {
		return 0, errors.New("not exist capacity " + s)
	}
	return util.ToUint64(string(resp.Kvs[0].Value)), nil
}

func (oc *ObjectCapacity) CurDiskInfo() (disk.Info, error) {
	return disk.GetInfo(`\`)
}
