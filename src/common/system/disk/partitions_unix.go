//go:build linux || netbsd || freebsd
// +build linux netbsd freebsd

package disk

import (
	"strings"

	"github.com/shirou/gopsutil/v3/disk"
)

// DeviceMountPoint returns the map of device name (without /dev/ prefix) to it's mount-points
func DeviceMountPoints() (map[string][]string, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}
	mp := make(map[string][]string, len(parts))
	for _, part := range parts {
		name := strings.TrimPrefix(part.Device, "/dev/")
		mp[name] = append(mp[name], part.Mountpoint)
	}
	return mp, nil
}

// MountPointDevice returns the map of mount-point to it's device name (without /dev/ prefix)
func MountPointDevice() (map[string]string, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}
	mp := make(map[string]string, len(parts))
	for _, part := range parts {
		mp[part.Mountpoint] = strings.TrimPrefix(part.Device, "/dev/")
	}
	return mp, nil
}