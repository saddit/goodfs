package disk

import (
	"common/datasize"
	"fmt"
	"testing"
)

func TestGetInfo(t *testing.T) {
	info, err := GetInfo(`\`)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", info)
	total := datasize.MustParse(fmt.Sprintf("%dB", info.Total))
	free := datasize.MustParse(fmt.Sprintf("%dB", info.Free))
	t.Logf("total=%.1fGB, free=%.1fGB", total.GigaByte(), free.GigaByte())
}
