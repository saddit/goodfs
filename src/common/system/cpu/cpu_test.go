package cpu

import (
	"common/util"
	"testing"
)

func TestGetStat(t *testing.T) {
	stat, _ := StatInfo()
	bt, _ := util.EncodeMsgp(stat)
	var info Stat
	util.DecodeMsgp(&info, bt)
	t.Logf("%+v", info)
}