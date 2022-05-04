package temp

import (
	"github.com/allegro/bigcache"
	"goodfs/lib/util/cache"
	"goodfs/objectserver/model"
	"testing"
)

func TestCache(t *testing.T) {
	ca := cache.NewCache(bigcache.DefaultConfig(10))
	if ok := ca.SetGob("test", model.TempInfo{
		Name: "TestName",
		Id:   "TestId",
		Size: 0,
	}); ok {
		if res, ok := cache.GetGob[model.TempInfo](ca, "test"); ok {
			t.Logf("%v", res)
			return
		}
		t.Error("Get cache error")
	}
	t.Error("Put cache error")
}

func TestNilSlice(t *testing.T) {
	var s []int
	if s == nil {
		t.Logf("OK")
	} else {
		t.Error("No")
	}
}
