package system

import (
	"common/graceful"
	"common/logs"
	"common/system/disk"
	"common/util"
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type StatSyncer struct {
	cli     clientv3.KV
	key     string
	LeaseID clientv3.LeaseID
}

func Syncer(c clientv3.KV, key string) *StatSyncer {
	return &StatSyncer{cli: c, key: key}
}

func (d *StatSyncer) StartAutoSave() func() {
	ctx, cancel := context.WithCancel(context.Background())
	tk := time.NewTicker(time.Minute)
	go func() {
		defer graceful.Recover()
		for {
			select {
			case <-ctx.Done():
				logs.Std().Info("stop sync sys-info")
				return
			case <-tk.C:
				util.LogErrWithPre("sync sys-info", d.Sync())
			}
		}
	}()
	return func() {
		cancel()
		util.LogErrWithPre("remove sys-info", d.Clear())
	}
}

func (d *StatSyncer) Sync() error {
	info, err := NewInfo(disk.Root)
	if err != nil {
		return err
	}
	bt, err := util.EncodeMsgp(info)
	if err != nil {
		return err
	}
	_, err = d.cli.Put(context.Background(), d.key, string(bt), clientv3.WithLease(d.LeaseID))
	return err
}

func (d *StatSyncer) Clear() error {
	_, err := d.cli.Delete(context.Background(), d.key)
	return err
}
