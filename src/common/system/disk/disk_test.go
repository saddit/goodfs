package disk

import (
	"common/logs"
	"testing"

	"github.com/shirou/gopsutil/v3/disk"
)

func init() {
	logs.SetLevel(logs.Debug)
}

func TestGetInfo(t *testing.T) {
	info, err := GetInfo(Root)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("total=%dGB, free=%dGB, used=%dGB", info.Total.GigaByte(), info.Free.GigaByte(), info.Used.GigaByte())
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
