package disk

import (
	"common/logs"
	"github.com/shirou/gopsutil/v3/disk"
	"testing"
)

func init() {
	logs.SetLevel(logs.Debug)
}

func TestGetInfo(t *testing.T) {
	info, err := GetInfo(Root)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("total=%dGB, free=%dGB", info.Total.GigaByte(), info.Total.GigaByte())
}

func TestAllMountPoints(t *testing.T) {
	paths, err := AllMountPoints()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", paths)
}

func TestIOCounter(t *testing.T) {
	d, _ := disk.IOCounters()
	for _, stat := range d {
		t.Logf("%+v", stat)
	}
}
