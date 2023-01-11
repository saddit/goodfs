package service

import (
	"common/cst"
	"common/graceful"
	"common/response"
	"common/system/disk"
	"common/util"
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
	size, err := WriteFile(filepath.Join(global.Config.StoragePath, fileName), fileStream)
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

func Get(name string, offset, size int64, writer io.Writer) error {
	if err := GetFile(filepath.Join(global.Config.StoragePath, name), offset, size, writer); err != nil {
		return err
	}
	MarkExist(name)
	return nil
}

func GetTemp(name string, size int64, writer io.Writer) error {
	return GetFile(filepath.Join(global.Config.TempPath, name), 0, size, writer)
}

func GetFile(fullPath string, offset, size int64, writer io.Writer) error {
	f, err := disk.OpenFileDirectIO(fullPath, os.O_RDONLY, cst.OS.ModeUser)
	if util.IsOSNotExist(err) {
		return response.NewError(404, "object not found")
	}
	if err != nil {
		return err
	}
	defer f.Close()
	if offset > 0 {
		if int(offset)%cst.OS.PageSize > 0 {
			return response.NewError(400, "offset must be power of 4KB")
		}
		if _, err = f.Seek(offset, io.SeekCurrent); err != nil {
			return err
		}
	}
	if _, err = io.CopyBuffer(disk.LimitWriter(writer, size), f, disk.AlignedBlock(8*cst.OS.PageSize)); err != nil {
		return err
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
	if util.IsOSNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if err = os.Remove(pt); err != nil && !util.IsOSNotExist(err) {
		return 0, err
	}
	return info.Size(), nil
}

// WriteFile 如果连续写入1次以上不满足4KB倍数的数据，中间将会产生无效padding，读取时无法去除文件中间的padding
func WriteFile(fullPath string, fileStream io.Reader) (int64, error) {
	file, err := disk.OpenFileDirectIO(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, cst.OS.ModeUser)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	// write file and aligen to power of 4KB
	return disk.AligendWriteTo(file, fileStream, 8*cst.OS.PageSize)
}

func MvTmpToStorage(tmpName, fileName string) error {
	filePath := filepath.Join(global.Config.StoragePath, fileName)
	tempPath := filepath.Join(global.Config.TempPath, tmpName)
	if ExistPath(filePath) {
		return nil
	}
	if err := os.Rename(tempPath, filePath); err != nil {
		if os.IsNotExist(err) {
			return response.NewError(404, "object not found")
		}
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
