package service

import (
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

func Put(fileName string, fileStream io.Reader) error {
	return AppendFile(global.Config.StoragePath, fileName, fileStream)
}

func Get(name string, writer io.Writer) error {
	return GetFile(filepath.Join(global.Config.StoragePath, name), writer)
}

func GetTemp(name string, writer io.Writer) error {
	return GetFile(filepath.Join(global.Config.TempPath, name), writer)
}

func GetFile(fullPath string, writer io.Writer) error {
	f, e := os.Open(fullPath)
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
	e := os.Remove(filepath.Join(path, name))
	if e != nil {
		return e
	}
	return nil
}

func AppendFile(path, fileName string, fileStream io.Reader) error {
	path = filepath.Join(path, fileName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND, os.ModePerm)
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
	filePath := filepath.Join(global.Config.StoragePath, fileName)
	tempPath := filepath.Join(global.Config.TempPath, tmpName)
	if ExistPath(filePath) {
		return nil
	}
	return os.Rename(tempPath, filePath)
}
