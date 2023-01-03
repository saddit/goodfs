package service

import (
	"apiserver/config"
	"common/util"
	"fmt"
	"go.uber.org/atomic"
	"io"
)

type CopyPutStream struct {
	rpConfig config.ReplicationConfig
	writers  []io.WriteCloser
}

func NewCopyPutStream(hash string, size int64, ips []string, rpCfg *config.ReplicationConfig) (*CopyPutStream, error) {
	writers := make([]io.WriteCloser, rpCfg.CopiesCount)
	wg := util.NewDoneGroup()
	defer wg.Close()
	for i := range writers {
		wg.Todo()
		go func(idx int) {
			defer wg.Done()
			stream, e := NewPutStream(ips[idx], fmt.Sprintf("%s.%d", hash, idx), size)
			if e != nil {
				wg.Error(e)
			} else {
				writers[idx] = stream
			}
		}(i)
	}
	if err := wg.WaitUntilError(); err != nil {
		return nil, err
	}
	return &CopyPutStream{
		rpConfig: *rpCfg,
		writers:  writers,
	}, nil
}

func (c *CopyPutStream) Write(p []byte) (n int, err error) {
	dg := util.NewDoneGroup()
	defer dg.Close()
	sucNum := atomic.NewInt32(0)
	for _, wt := range c.writers {
		dg.Todo()
		go func(writer io.Writer) {
			defer dg.Done()
			if _, err := writer.Write(p); err != nil {
				dg.Error(err)
				return
			}
			sucNum.Inc()
		}(wt)
	}
	if err := dg.WaitUntilError(); err != nil && c.rpConfig.AtLeastCopiesNum() > int(sucNum.Load()) {
		return 0, err
	}
	return len(p), nil
}

func (c *CopyPutStream) Close() error {
	var err error
	for _, wt := range c.writers {
		if inner := wt.Close(); inner != nil {
			err = inner
		}
	}
	return err
}

func (c *CopyPutStream) Commit(ok bool) error {
	wg := util.NewDoneGroup()
	defer wg.Close()
	for _, w := range c.writers {
		if util.InstanceOf[Committer](w) {
			wg.Todo()
			go func(cm Committer) {
				defer wg.Done()
				if e := cm.Commit(ok); e != nil {
					wg.Error(e)
				}
			}(w.(Committer))
		}
	}
	return wg.WaitUntilError()
}
