package service

import (
	"goodfs/objectserver/global"
	"io"
	"os"
)

const (
	LocateKeyPrefix = "LocateCache#"
)

func Exist(name string) bool {
	if global.Cache.Has(LocateKeyPrefix+name) || global.Cache.Has(name) {
		return true
	}
	if ExistPath(global.Config.StoragePath + name) {
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

func Put(fileName string, fileStream io.Reader) error {
	return PutFile(global.Config.StoragePath, fileName, fileStream)
}

func Get(name string, writer io.Writer) error {
	f, e := os.Open(global.Config.StoragePath + name)
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
	return DeleteFile(global.Config.StoragePath, name)
}

func DeleteFile(path, name string) error {
	e := os.Remove(path + name)
	if e != nil {
		return e
	}
	return nil
}

func PutFile(path, fileName string, fileStream io.Reader) error {
	file, err := os.Create(path + fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = io.Copy(file, fileStream); err != nil {
		return err
	}
	MarkExist(fileName)
	return nil
}

func MvTmpToStorage(tmpName, fileName string) error {
	return os.Rename(global.Config.TempPath+tmpName, global.Config.StoragePath+fileName)
}
