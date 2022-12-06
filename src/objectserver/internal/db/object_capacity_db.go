package db

import (
	"common/cst"
	"common/graceful"
	"common/logs"
	"common/system"
	"common/system/disk"
	"common/util"
	"context"
	"errors"
	"objectserver/config"
	"strings"
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
	util.LogErrWithPre("auto-save object-cap and sys-info", oc.Save())
	go func() {
		defer graceful.Recover()
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				util.LogErrWithPre("auto-save object-cap and sys-info", oc.Save())
			case <-ctx.Done():
				logs.Std().Info("stop auto-save object-cap and sys-info")
				return
			}
		}
	}()
	return func() {
		cancel()
		util.LogErrWithPre("remove capacity", oc.RemoveAll())
	}
}

func (oc *ObjectCapacity) Save() error {
	dg := util.NewDoneGroup()
	defer dg.Close()
	// save object-cap
	dg.Todo()
	go func() {
		defer dg.Done()
		keyCap := cst.EtcdPrefix.FmtObjectCap(oc.groupName, oc.serviceName, oc.CurrentID)
		if _, err := oc.cli.Put(context.Background(), keyCap, oc.CurrentCap.String()); err != nil {
			dg.Error(err)
			return
		}
	}()
	// save disk-info
	dg.Todo()
	go func() {
		defer dg.Done()
		info, err := system.NewInfo(disk.Root)
		if err != nil {
			dg.Error(err)
			return
		}
		bt, err := util.EncodeMsgp(info)
		if err != nil {
			dg.Error(err)
			return
		}
		keyDisk := cst.EtcdPrefix.FmtSystemInfo(oc.groupName, oc.serviceName, oc.CurrentID)
		if _, err = oc.cli.Put(context.Background(), keyDisk, string(bt)); err != nil {
			dg.Error(err)
			return
		}
	}()
	return dg.WaitUntilError()
}

func (oc *ObjectCapacity) GetAll() (map[string]uint64, error) {
	resp, err := oc.cli.Get(context.Background(), cst.EtcdPrefix.ObjectCap, clientv3.WithPrefix())
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
	key := cst.EtcdPrefix.FmtObjectCap(oc.groupName, oc.serviceName, oc.CurrentID)
	resp, err := oc.cli.Get(context.Background(), key)
	if err != nil {
		return 0, err
	}
	if len(resp.Kvs) == 0 {
		return 0, errors.New("not exist capacity " + s)
	}
	return util.ToUint64(string(resp.Kvs[0].Value)), nil
}

func (oc *ObjectCapacity) RemoveAll() error {
	keyDisk := cst.EtcdPrefix.FmtSystemInfo(oc.groupName, oc.serviceName, oc.CurrentID)
	_, err := oc.cli.Delete(context.Background(), keyDisk)
	if err != nil {
		return err
	}
	keyCap := cst.EtcdPrefix.FmtObjectCap(oc.groupName, oc.serviceName, oc.CurrentID)
	_, err = oc.cli.Delete(context.Background(), keyCap)
	if err != nil {
		return err
	}
	return nil
}
