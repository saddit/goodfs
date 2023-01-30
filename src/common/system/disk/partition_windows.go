//go:build windows

package disk

import (
	"github.com/shirou/gopsutil/v3/disk"
)

// DeviceMountPoint returns the map of device (volume) name to it's mount-point
func DeviceMountPoints() (map[string][]string, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}
	mp := make(map[string][]string, len(parts))
	for _, part := range parts {
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
		mp[part.Mountpoint] = part.Device
	}
	return mp, nil
}