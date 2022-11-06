package logic

import (
	"apiserver/internal/usecase/pool"
	"common/constrant"
	"common/graceful"
	"common/logs"
	"common/system"
	"common/system/cpu"
	"common/system/disk"
	"common/system/mem"
	"common/util"
	"context"
	"time"
)

type SystemStatLogic struct {
}

func NewSystemStatLogic() *SystemStatLogic {
	return &SystemStatLogic{}
}

func (d SystemStatLogic) StartAutoSave() func() {
	ctx, cancel := context.WithCancel(context.Background())
	tk := time.NewTicker(time.Minute)
	util.LogErrWithPre("auto save sys-info", d.Save())
	go func() {
		defer graceful.Recover()
		for {
			select {
			case <-ctx.Done():
				logs.Std().Info("stop auto save sys-info")
				return
			case <-tk.C:
				util.LogErrWithPre("auto save sys-info", d.Save())
			}
		}
	}()
	return cancel
}

func (d SystemStatLogic) Save() error {
	cpuStat, err := cpu.StatInfo()
	if err != nil {
		return err
	}
	memStat, err := mem.MemStat()
	if err != nil {
		return err
	}
	bt, err := util.EncodeMsgp(&system.Info{
		DiskInfo:  disk.Info{},
		MemStatus: memStat,
		CpuStatus: cpuStat,
	})
	if err != nil {
		return err
	}
	keyDisk := constrant.EtcdPrefix.FmtSystemInfo(pool.Config.Registry.Group, pool.Config.Registry.Name, pool.Config.Registry.ServerID)
	_, err = pool.Etcd.Put(context.Background(), keyDisk, string(bt))
	return err
}
