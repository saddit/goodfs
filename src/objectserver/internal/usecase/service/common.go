package service

import (
	"common/cache"
	"common/graceful"
	"common/logs"
	"common/util"
	"context"
	"io/fs"
	"objectserver/internal/entity"
	"objectserver/internal/usecase/pool"
	"path/filepath"
	"strings"
)

var log = logs.New("locate-service")

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
			log.Error(err)
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
		log.Infof("cleaner[%d] start handling temp file removal..", tid)
		for {
			select {
			case entry := <-ch:
				log.Debugf("cleaner[%d] cache key %s evicted", tid, entry.Key)
				if strings.HasPrefix(entry.Key, entity.TempKeyPrefix) {
					var ti entity.TempInfo
					if ok := util.GobDecode(entry.Value, &ti); ok {
						if _, err := DeleteFile(pool.Config.TempPath, ti.Id); err != nil {
							log.Errorf("cleaner[%d] remove temp %v(name=%v) error, %v", tid, ti.Id, ti.Name, err)
							break
						}
						log.Debugf("cleaner[%d] remove temp file %s", tid, ti.Id)
					} else {
						log.Errorf("cleaner[%d] handle evicted key=%v error, value cannot cast to TempInfo", tid, entry.Key)
					}
				}
			case <-ctx.Done():
				log.Infof("cleaner[%d] stop handling temp file removal", tid)
				return
			}
		}
	}

	for i := 0; i < threadNum; i++ {
		go startCleaner(i + 1)
	}
	return cancel
}
