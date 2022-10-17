package service

import (
	"common/constrant"
	"common/system/disk"
	"common/graceful"
	"io"
	global "objectserver/internal/usecase/pool"
	"os"
	"path/filepath"
)

const (
	LocateKeyPrefix = "LocateCache#"
)

func Exist(name string) bool {
	if global.Cache.Has(LocateKeyPrefix+name) || global.Cache.Has(name) {
		return true
	}
	if ExistPath(filepath.Join(global.Config.StoragePath, name)) {
		MarkExist(name)
		return true
	} else {
		return false
	}
}

func ExistPath(fullPath string) bool {
	_, err := os.Stat(fullPath)
	return os.IsExist(err) || !os.IsNotExist(err)
}

func MarkExist(name string) {
	global.Cache.Set(LocateKeyPrefix+name, []byte{})
}

func UnMarkExist(name string) {
	global.Cache.Delete(LocateKeyPrefix + name)
}

func Put(fileName string, fileStream io.Reader) error {
	if Exist(fileName) {
		return nil
	}
	size, err := AppendFile(global.Config.StoragePath, fileName, fileStream)
	if err != nil {
		return err
	}
	go func() {
		defer graceful.Recover()
		global.ObjectCap.CurrentCap.Add(uint64(size))
		MarkExist(fileName)
	}()
	return nil
}

func Get(name string, writer io.Writer) error {
	if err := GetFile(filepath.Join(global.Config.StoragePath, name), writer); err != nil {
		return err
	}
	MarkExist(name)
	return nil
}

func GetTemp(name string, writer io.Writer) error {
	return GetFile(filepath.Join(global.Config.TempPath, name), writer)
}

func GetFile(fullPath string, writer io.Writer) error {
	f, e := disk.OpenFileDirectIO(fullPath, os.O_RDONLY, 0)
	if e != nil {
		return e
	}
	defer f.Close()
	if _, e = io.Copy(writer, f); e != nil {
		return e
	}
	return nil
}

func Delete(name string) error {
	size, err := DeleteFile(global.Config.StoragePath, name)
	if err != nil {
		return err
	}
	go func() {
		defer graceful.Recover()
		global.ObjectCap.CurrentCap.Sub(uint64(size))
		UnMarkExist(name)
	}()
	return nil
}

func DeleteFile(path, name string) (int64, error) {
	pt := filepath.Join(path, name)
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), os.Remove(pt)
}

func AppendFile(path, fileName string, fileStream io.Reader) (int64, error) {
	path = filepath.Join(path, fileName)
	file, err := disk.OpenFileDirectIO(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, constrant.OS.ModeUser)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return io.Copy(file, fileStream)
}

func MvTmpToStorage(tmpName, fileName string) error {
	filePath := filepath.Join(global.Config.StoragePath, fileName)
	tempPath := filepath.Join(global.Config.TempPath, tmpName)
	if ExistPath(filePath) {
		return nil
	}
	if err := os.Rename(tempPath, filePath); err != nil {
		return err
	}
	go func() {
		defer graceful.Recover()
		if info, err := os.Stat(filePath); err == nil {
			global.ObjectCap.CurrentCap.Add(uint64(info.Size()))
		}
		MarkExist(fileName)
	}()
	return nil
}
