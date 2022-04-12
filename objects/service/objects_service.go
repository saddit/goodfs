package service

import (
	"goodfs/objects/config"
	"io"
	"os"
)

func Exist(name string) bool {
	_, err := os.Stat(config.StoragePath + name)
	return !os.IsNotExist(err)
}

func Put(fileName string, fileStream io.Reader) error {
	file, err := os.Create(config.StoragePath + fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	io.Copy(file, fileStream)
	return nil
}

func Get(name string, writer io.Writer) error {
	f, e := os.Open(config.StoragePath + name)
	if e != nil {
		return e
	}
	defer f.Close()
	io.Copy(writer, f)
	return nil
}

func Delete(name string) error {
	e := os.Remove(config.StoragePath + name)
	if e != nil {
		return e
	}
	return nil
}
