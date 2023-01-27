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
			pool.ObjectCap.AddCap(info.Size())
			if !info.IsDir() {
				MarkExist(info.Name())
			}
			return nil
		})
		if err != nil {
			log.Error(err)
			return
		}
	}
	if err := pool.ObjectCap.Save(); err != nil {
		log.Error(err)
		return
	}
}

// StartTempRemovalBackground start a temp file removal thread. watching the eviction of cache.
// return cancel function.
func StartTempRemovalBackground(cache cache.ICache) func() {
	ch := cache.NotifyEvicted()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer graceful.Recover()
		log.Info("Start handling temp file removal..")
		for {
			select {
			case entry := <-ch:
				log.Debugf("cache key %s evicted", entry.Key)
				if strings.HasPrefix(entry.Key, entity.TempKeyPrefix) {
					var ti entity.TempInfo
					if ok := util.GobDecode(entry.Value, &ti); ok {
						if _, err := DeleteFile(pool.Config.TempPath, ti.Id); err != nil {
							log.Errorf("Remove temp %v(name=%v) error, %v", ti.Id, ti.Name, err)
							break
						}
						log.Debugf("Remove temp file %s", ti.Id)
					} else {
						log.Errorf("Handle evicted key=%v error, value cannot cast to TempInfo", entry.Key)
					}
				}
			case <-ctx.Done():
				log.Info("Stop handling temp file removal")
				return
			}
		}
	}()
	return cancel
}
