package service

import (
	"common/logs"
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
