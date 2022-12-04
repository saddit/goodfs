package disk

import (
	"testing"
)

func TestGetInfo(t *testing.T) {
	info, err := GetInfo(Root)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("total=%dGB, free=%dGB", info.Total.GigaByte(), info.Total.GigaByte())
}
