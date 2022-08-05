package service

import (
	"common/graceful"
	"common/logs"
	"common/util"
	"objectserver/internal/entity"
	"objectserver/internal/usecase/pool"
	"os"
	"strings"
)

var logrus = logs.New("locate-service")

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

func StartTempRemovalBackground() {
	go func ()  {
		defer graceful.Recover()
		ch := pool.Cache.NotifyEvicted()
		logrus.Info("Start handle temp file removal..")
		for entry := range ch {
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
		}
	}()
}
