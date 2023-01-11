package service

import (
	"apiserver/config"
	"apiserver/internal/usecase/webapi"
	"bytes"
	"common/graceful"
	"common/util"
	"errors"
	"fmt"
	"strings"
)

type CopyFixStream struct {
	fileNames []string
	locates   []string
	buffer    *bytes.Buffer
	rpConfig  *config.ReplicationConfig
	Updater   LocatesUpdater
}

func NewCopyFixStream(lostNames []string, newLocates []string, up LocatesUpdater, cfg *config.ReplicationConfig) (*CopyFixStream, error) {
	return &CopyFixStream{
		fileNames: lostNames,
		locates:   newLocates,
		rpConfig:  cfg,
		buffer:    bytes.NewBuffer(make([]byte, 0, 4096)),
		Updater:   up,
	}, nil
}

func (c *CopyFixStream) Write(bt []byte) (int, error) {
	return c.buffer.Write(bt)
}

func (c *CopyFixStream) Close() error {
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
		defer dg.Done()
		var errs []string
		data := c.buffer.Bytes()
		for idx, name := range c.fileNames {
			if err := webapi.PutObject(c.locates[idx], name, bytes.NewBuffer(data)); err != nil {
				errs = append(errs, fmt.Sprintf("fix %s put-api err: %s", name, err))
			}
		}
		if len(errs) > 0 {
			dg.Error(errors.New(strings.Join(errs, ";")))
			return
		}
		if c.Updater == nil {
			dg.Error(errors.New("locates updater required but nil"))
			return
		}
		if err := c.Updater(c.locates); err != nil {
			dg.Error(err)
			return
		}
	}()
	return dg.WaitUntilError
}
