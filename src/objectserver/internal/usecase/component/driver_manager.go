package component

import (
	"common/collection/set"
	"common/datasize"
	"common/graceful"
	"common/logs"
	"common/system/disk"
	"common/util"
	"context"
	"os"
	"path/filepath"
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
	Excludes set.Set
}

func NewDriverManager(lb DriverBalancer, excludes ...string) *DriverManager {
	return &DriverManager{balancer: lb, Excludes: set.OfString(excludes)}
}

func (dm *DriverManager) Update() {
	mps, err := disk.AllMountPoints()
	if err != nil {
		logs.Std().Errorf("update driver info err: %s", err)
		return
	}
	info := make([]*Driver, 0, len(mps))
	for _, mp := range mps {
		// skip excluded mount point
		if dm.Excludes.Contains(mp) {
			continue
		}
		// get driver stat
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

func (dm *DriverManager) StartAutoUpdate() (stop func()) {
	ctx, cancel := context.WithCancel(context.Background())
	stop = cancel
	tk := util.ImmediateTick(time.Minute)
	go func() {
		defer graceful.Recover()
		for {
			select {
			case <-ctx.Done():
				logs.Std().Info("stop update drivers")
				return
			case <-tk:
				dm.Update()
			}
		}
	}()
	return
}

func (dm *DriverManager) SelectDriver() (*Driver, error) {
	return dm.balancer.Select(dm.drivers)
}

func (dm *DriverManager) SelectMountPointFallback(fb string) string {
	d, err := dm.balancer.Select(dm.drivers)
	if err != nil {
		return fb
	}
	return d.MountPoint
}

func (dm *DriverManager) IsPathExist(path string) bool {
	for _, d := range dm.drivers {
		if _, err := os.Stat(filepath.Join(d.MountPoint, path)); !os.IsNotExist(err) {
			return true
		}
	}
	return false
}

func (dm *DriverManager) FindMountPath(path string) (string, error) {
	for _, d := range dm.drivers {
		fullPath := filepath.Join(d.MountPoint, path)
		if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
			return path, nil
		}
	}
	return "", os.ErrNotExist
}

func (dm *DriverManager) GetAllMountPoint() []string {
	res := make([]string, 0, len(dm.drivers))
	for _, driver := range dm.drivers {
		res = append(res, driver.MountPoint)
	}
	return res
}
