package service

import (
	"apiserver/config"
	"apiserver/internal/usecase/webapi"
	"bufio"
	"common/graceful"
	"common/util"
	"errors"
	"fmt"
	"os"
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
	tmp, err := os.CreateTemp("", lostNames[0])
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
	if err := c.fixFile.Close(); err != nil {
		return err
	}
	//FIXME: 写入的文件比原始文件大
	wait := c.startFix()
	if c.rpConfig.CopyAsync {
		go func() {
			defer graceful.Recover()
			util.LogErrWithPre("copy-fix-stream", wait())
		}()
		return nil
	}
	return wait()
}

func (c *CopyFixStream) startFix() func() error {
	dg := util.NewDoneGroup()
	dg.Todo()
	go func() {
		defer func() { util.LogErr(os.Remove(c.fixFile.Name())) }()
		defer dg.Done()
		var errs []string
		for idx, name := range c.fileNames {
			file, err := os.Open(c.fixFile.Name())
			if err != nil {
				errs = append(errs, fmt.Sprintf("open file err: %s", err))
				continue
			}
			if err = webapi.PutObject(c.locates[idx], name, file); err != nil {
				errs = append(errs, fmt.Sprintf("fix %s put-api err: %s", name, err))
			}
			_ = file.Close()
		}
		if len(errs) > 0 {
			dg.Error(errors.New(strings.Join(errs, ";")))
		}
	}()
	return dg.WaitUntilError
}
