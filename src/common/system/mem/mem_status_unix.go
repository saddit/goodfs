//go:build linux || darwin
// +build linux darwin

package mem

import (
	"runtime"
	"syscall"
)

func MemStat() Status {
	// program memory usage
	memStat := new(runtime.MemStats)
	runtime.ReadMemStats(memStat)
	mem := Status{}
	mem.Self = memStat.Alloc

	// system memory usage
	sysInfo := new(syscall.Sysinfo_t)
	err := syscall.Sysinfo(sysInfo)
	if err == nil {
		mem.All = sysInfo.Totalram * uint64(syscall.Getpagesize())
		mem.Free = sysInfo.Freeram * uint64(syscall.Getpagesize())
		mem.Used = mem.All - mem.Free
	}
	return mem
}
