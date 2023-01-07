package disk

import (
	"common/logs"
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

func TestAllPartitionPath(t *testing.T) {
	paths, err := AllPartitionPath()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", paths)
}
