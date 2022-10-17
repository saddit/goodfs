package logic

import (
	"common/constrant"
	"common/system/disk"
	"common/graceful"
	"common/logs"
	"common/util"
	"context"
	"metaserver/internal/usecase/pool"
	"time"
)

type DiskLogic struct {
}

func NewDiskLogic() *DiskLogic {
	return &DiskLogic{}
}

func (d DiskLogic) StartAutoSave() func() {
	ctx, cancel := context.WithCancel(context.Background())
	tk := time.NewTicker(time.Minute)
	go func() {
		defer graceful.Recover()
		for {
			select {
			case <-ctx.Done():
				logs.Std().Info("stop auto save disk-info")
				return
			case <-tk.C:
				util.LogErrWithPre("auto save disk-info", d.Save())
			}
		}
	}()
	return cancel
}

func (d DiskLogic) Save() error {
	info, err := d.CurDiskInfo()
	if err != nil {
		return err
	}
	bt, err := util.EncodeMsgp(&info)
	if err != nil {
		return err
	}
	keyDisk := constrant.EtcdPrefix.FmtDiskInfo(pool.Config.Registry.Group, pool.Config.Registry.Name, pool.Config.Registry.ServerID)
	_, err = pool.Etcd.Put(context.Background(), keyDisk, string(bt))
	return err
}

func (DiskLogic) CurDiskInfo() (disk.Info, error) {
	return disk.GetInfo(`\`)
}
