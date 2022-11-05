//go:build windows

package mem

import (
	"github.com/shirou/gopsutil/v3/mem"
)

func MemStat() (res Status, err error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	res.All = stat.Total
	res.Used = stat.Used
	res.Free = stat.Free
	return
}
