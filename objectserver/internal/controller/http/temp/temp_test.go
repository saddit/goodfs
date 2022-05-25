package temp

import (
	"common/cache"
	"objectserver/internal/entity"
	"testing"

	"github.com/allegro/bigcache"
)

func TestCache(t *testing.T) {
	ca := cache.NewCache(bigcache.DefaultConfig(10))
	if ok := ca.SetGob("test", entity.TempInfo{
		Name: "TestName",
		Id:   "TestId",
		Size: 0,
	}); ok {
		if res, ok := cache.GetGob[entity.TempInfo](ca, "test"); ok {
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
