package service

import (
	"goodfs/objectserver/config"
	"io"
	"os"
)

func Exist(name string) bool {
	_, err := os.Stat(config.StoragePath + name)
	return !os.IsNotExist(err)
}

func Put(fileName string, fileStream io.Reader) error {
	return PutFile(config.StoragePath, fileName, fileStream)
}

func Get(name string, writer io.Writer) error {
	f, e := os.Open(config.StoragePath + name)
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
	e := os.Remove(config.StoragePath + name)
	if e != nil {
		return e
	}
	return nil
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
	return nil
}

func MvTmpToStorage(tmpName, fileName string) error {
	return os.Rename(config.TempPath+tmpName, config.StoragePath+fileName)
}
