package service

import (
	"common/cache"
	"common/graceful"
	"common/logs"
	"common/util"
	"context"
	"fmt"
	"io/fs"
	"objectserver/internal/entity"
	"objectserver/internal/usecase/pool"
	"path/filepath"
	"strings"
)

// WarmUpLocateCache walk all objects under the storage path and save marks to cache
func WarmUpLocateCache() {
	for _, mp := range pool.DriverManager.GetAllMountPoint() {
		err := filepath.Walk(filepath.Join(mp, pool.Config.StoragePath), func(_ string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fileName := info.Name()
			// skip temp file
			if strings.HasPrefix(fileName, entity.TempKeyPrefix) {
				return nil
			}
			pool.ObjectCap.AddCap(info.Size())
			if !info.IsDir() {
				MarkExist(fileName)
			}
			return nil
		})
		if err != nil {
			logs.Std().Error(err)
			return
		}
	}
}

// StartTempRemovalBackground start some temp file removal threads. watching the eviction of cache.
// return cancel function.
func StartTempRemovalBackground(cache cache.ICache, threadNum int) func() {
	ch := cache.NotifyEvicted()
	ctx, cancel := context.WithCancel(context.Background())
	startCleaner := func(tid int) {
		defer graceful.Recover()
		logger := logs.New(fmt.Sprintf("cleaner-%d", tid))
		logger.Info("start handling temp file removal..")
		for {
			select {
			case entry := <-ch:
				logger.Debugf("cache key %s evicted", entry.Key)
				if strings.HasPrefix(entry.Key, entity.TempKeyPrefix) {
					var ti entity.TempInfo
					if ok := util.GobDecode(entry.Value, &ti); ok {
						if _, err := DeleteFile(pool.Config.TempPath, ti.Id); err != nil {
							logger.Errorf("remove temp %v(name=%v) error, %v", ti.Id, ti.Name, err)
							break
						}
						logger.Debugf("remove temp file %s", ti.Id)
					} else {
						logger.Errorf("handle evicted key=%v error, value cannot cast to TempInfo", entry.Key)
					}
				}
			case <-ctx.Done():
				logger.Info("stop handling temp file removal")
				return
			}
		}
	}

	for i := 0; i < threadNum; i++ {
		go startCleaner(i + 1)
	}
	return cancel
}
