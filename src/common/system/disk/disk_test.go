package disk

import (
	"testing"
)

func TestGetInfo(t *testing.T) {
	info, err := GetInfo(Root)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("total=%.1fGB, free=%.1fGB", info.Total.GigaByte(), info.Total.GigaByte())
}
