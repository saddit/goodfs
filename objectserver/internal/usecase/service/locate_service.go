package service

import (
	"common/cache"
	"common/graceful"
	"common/logs"
	"common/util"
	"context"
	"github.com/sirupsen/logrus"
	"objectserver/internal/entity"
	"objectserver/internal/usecase/pool"
	"os"
	"strings"
)

var log = logs.New("locate-service")

func WarmUpLocateCache() {
	files, e := os.ReadDir(pool.Config.StoragePath)
	if e != nil {
		panic(e)
	}
	for _, f := range files {
		if !f.IsDir() {
			MarkExist(f.Name())
		}
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
						if e := DeleteFile(pool.Config.TempPath, ti.Id); e != nil {
							logrus.Infof("Remove temp %v(name=%v) error, %v", ti.Id, ti.Name, e)
						}
					} else {
						logrus.Infof("Handle evicted key=%v error, value cannot cast to TempInfo", entry.Key)
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
