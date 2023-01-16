package disk

import (
	"github.com/shirou/gopsutil/v3/disk"
)

func AllMountPoints() ([]string, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, part := range parts {
		paths = append(paths, part.Mountpoint)
	}
	return paths, nil
}
