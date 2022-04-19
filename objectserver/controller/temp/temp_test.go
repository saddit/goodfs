package temp

import (
	"github.com/allegro/bigcache"
	"goodfs/objectserver/config"
	"goodfs/objectserver/model"
	"goodfs/util/cache"
	"testing"
)

func TestCache(t *testing.T) {
	ca := cache.NewCache(bigcache.DefaultConfig(config.CacheTTL))
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
