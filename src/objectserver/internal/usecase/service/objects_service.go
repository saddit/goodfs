package service

import (
	"bytes"
	"common/cst"
	"common/graceful"
	"common/logs"
	"common/response"
	"common/system/disk"
	"fmt"
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
	if os.IsNotExist(err) {
		return response.NewError(404, "object not found")
	}
	if err != nil {
		return err
	}
	defer f.Close()
	if offset > 0 {
		if int(offset)%cst.OS.PageSize > 0 {
			logs.Std().Warn("offset must be power of 4KB, direct-io will disabled")
			if err = disk.DisableDirectIO(f); err != nil {
				return fmt.Errorf("diable direct-io: %w", err)
			}
		}
		if _, err = f.Seek(offset, io.SeekStart); err != nil {
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
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if err = os.Remove(pt); err != nil && !os.IsNotExist(err) {
		return 0, err
	}
	return info.Size(), nil
}

// WriteFileWithSize will use provided curSize to remove padding of tail
// and keep writing data aligened to multiple of 4KB
func WriteFileWithSize(fullPath string, curSize int64, fileStream io.Reader) (int64, error) {
	file, err := disk.OpenFileDirectIO(fullPath, os.O_RDWR|os.O_CREATE, cst.OS.ModeUser)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return 0, err
	}
	// paddingLen always gte 0 and lt 4096
	pageSize := int64(cst.OS.PageSize)
	paddingLen := fi.Size() - curSize
	if paddingLen >= pageSize {
		return 0, fmt.Errorf("err padding length %d", paddingLen)
	}
	if paddingLen > 0 {
		// read the last 4KB of data
		if _, err = file.Seek(-pageSize, io.SeekEnd); err != nil {
			return 0, err
		}
		bt := make([]byte, pageSize)
		n, err := io.ReadFull(file, bt)
		if err != nil {
			return 0, err
		}
		if int64(n) < pageSize {
			return 0, fmt.Errorf("read tail except %d but %d", pageSize, n)
		}
		// remove padding
		bt = bt[:len(bt) - int(paddingLen)]
		// concatenation with fileStream
		fileStream = disk.MultiReader(bytes.NewBuffer(bt), fileStream)
		// seek back
		if _, err = file.Seek(-pageSize, io.SeekEnd); err != nil {
			return 0, err
		}
	}
	// write file and aligen to power of 4KB
	return disk.NewAligendWriter(file).ReadFrom(fileStream)
}

// WriteFile should make sure size of each write is a multiple of 4096 (except last) 
func WriteFile(fullPath string, fileStream io.Reader) (int64, error) {
	file, err := disk.OpenFileDirectIO(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, cst.OS.ModeUser)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	// write file and aligen to power of 4KB
	return disk.NewAligendWriter(file).ReadFrom(fileStream)
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
