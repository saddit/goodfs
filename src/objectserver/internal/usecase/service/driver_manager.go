package service

import (
	"common/datasize"
	"common/graceful"
	"common/logs"
	"common/system/disk"
	"context"
	"time"
)

type Driver struct {
	MountPoint string
	FreeSpace  datasize.DataSize
	TotalSpace datasize.DataSize
}

type DriverManager struct {
	drivers  []*Driver
	balancer DriverBalancer
}

func NewDriverManager(lb DriverBalancer) *DriverManager {
	return &DriverManager{balancer: lb}
}

func (dm *DriverManager) StartAutoUpdate() (stop func()) {
	ctx, cancel := context.WithCancel(context.Background())
	stop = cancel
	tk := time.NewTicker(time.Minute)
	go func() {
		defer graceful.Recover()
		for {
			select {
			case <-ctx.Done():
				tk.Stop()
				logs.Std().Info("stop update drivers")
				return
			case <-tk.C:
				mps, err := disk.AllMountPoints()
				if err != nil {
					logs.Std().Errorf("update driver info err: %s", err)
					break
				}
				info := make([]*Driver, 0, len(mps))
				for _, mp := range mps {
					stat, err := disk.GetInfo(mp)
					if err != nil {
						logs.Std().Errorf("update driver '%s' err: %s", mp, err)
						continue
					}
					info = append(info, &Driver{
						MountPoint: mp,
						FreeSpace:  stat.Free,
						TotalSpace: stat.Total,
					})
				}
				dm.drivers = info
			}
		}
	}()
	return
}

func (dm *DriverManager) SelectDriver() (*Driver, error) {
	return dm.balancer.Select(dm.drivers)
}
