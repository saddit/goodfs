package mem

import (
	"github.com/shirou/gopsutil/v3/mem"
	"runtime"
)

func Stat() (res Status, err error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	res.All = stat.Total
	res.Used = stat.Used
	res.Free = stat.Free
	// program memory usage
	memStat := new(runtime.MemStats)
	runtime.ReadMemStats(memStat)
	res.Self = memStat.Alloc
	return
}
