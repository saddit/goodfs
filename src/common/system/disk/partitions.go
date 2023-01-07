package disk

import (
	"common/logs"
	"github.com/shirou/gopsutil/v3/disk"
)

func AllPartitionPath() ([]string, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, part := range parts {
		logs.Std().Debugf("device=%s,fsType=%s,mountPoint=%s", part.Device, part.Fstype, part.Mountpoint)
		paths = append(paths, part.Mountpoint)
	}
	return paths, nil
}
