package service

import (
	"apiserver/config"
	"apiserver/internal/usecase/webapi"
	"bufio"
	"common/system/disk"
	"common/util"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CopyFixStream struct {
	fileNames []string
	locates   []string
	writer    *bufio.Writer
	fixFile   *os.File
	rpConfig  *config.ReplicationConfig
}

func NewCopyFixStream(lostNames []string, newLocates []string, cfg *config.ReplicationConfig) (*CopyFixStream, error) {
	tmp, err := disk.OpenFileDirectIO(filepath.Join(os.TempDir(), lostNames[0]), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	return &CopyFixStream{
		fileNames: lostNames,
		locates:   newLocates,
		rpConfig:  cfg,
		writer:    bufio.NewWriter(tmp),
		fixFile:   tmp,
	}, nil
}

func (c *CopyFixStream) Write(bt []byte) (int, error) {
	return c.writer.Write(bt)
}

func (c *CopyFixStream) Close() error {
	if err := c.writer.Flush(); err != nil {
		return err
	}
	wait := c.startFix()
	if c.rpConfig.CopyAsync {
		return nil
	}
	return wait()
}

func (c *CopyFixStream) startFix() func() error {
	dg := util.NewDoneGroup()
	dg.Todo()
	go func() {
		defer func() { util.LogErr(os.Remove(c.fixFile.Name())) }()
		defer c.fixFile.Close()
		defer dg.Done()
		var errs []string
		for idx, name := range c.fileNames {
			if _, err := c.fixFile.Seek(0, io.SeekStart); err != nil {
				errs = append(errs, fmt.Sprintf("fix %s seek err: %s", name, err))
				continue
			}
			if err := webapi.PutObject(c.locates[idx], name, c.fixFile); err != nil {
				errs = append(errs, fmt.Sprintf("fix %s put-api err: %s", name, err))
			}
		}
		if len(errs) > 0 {
			dg.Error(errors.New(strings.Join(errs, ";")))
		}
	}()
	return dg.WaitUntilError
}
