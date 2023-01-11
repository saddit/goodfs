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

func WarmUpLocateCache() {
	err := filepath.Walk(pool.Config.StoragePath, func(_ string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		pool.ObjectCap.CurrentCap.Add(uint64(info.Size()))
		if !info.IsDir() {
			MarkExist(info.Name())
		}
		return nil
	})
	if err != nil {
		log.Error(err)
		return
	}
	if err = pool.ObjectCap.Save(); err != nil {
		log.Error(err)
		return
	}
}

// StartTempRemovalBackground 临时文件清除线程, 调用返回方法以停止
func StartTempRemovalBackground(cache cache.ICache) func() {
	ch := cache.NotifyEvicted()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer graceful.Recover()
		log.Info("Start handling temp file removal..")
		for {
			select {
			case entry := <-ch:
				if strings.HasPrefix(entry.Key, entity.TempKeyPrefix) {
					var ti entity.TempInfo
					if ok := util.GobDecodeGen2(entry.Value, &ti); ok {
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
