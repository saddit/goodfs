package service

import (
	"apiserver/config"
	"common/datasize"
	"common/util"
	"common/util/slices"
	"fmt"
	"io"
	"sync/atomic"
)

type CopyPutStream struct {
	rpConfig config.ReplicationConfig
	cache    []byte
	writers  []io.WriteCloser
}

func NewCopyPutStream(opt *StreamOption, rpCfg *config.ReplicationConfig) (*CopyPutStream, error) {
	writers := make([]io.WriteCloser, rpCfg.CopiesCount)
	wg := util.NewDoneGroup()
	defer wg.Close()
	for i := range writers {
		wg.Todo()
		go func(idx int) {
			defer wg.Done()
			stream, e := NewPutStream(opt.Locates[idx], fmt.Sprintf("%s.%d", opt.Hash, idx), opt.Size, opt.Compress)
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
		cache:    make([]byte, 0, rpCfg.BlockSize),
	}, nil
}

func (c *CopyPutStream) Flush() (err error) {
	if len(c.cache) == 0 {
		return nil
	}
	defer func() { slices.Clear(&c.cache) }()
	dg := util.NewDoneGroup()
	sucNum := atomic.Int32{}
	for _, wt := range c.writers {
		dg.Todo()
		go func(writer io.Writer) {
			defer dg.Done()
			if _, inner := writer.Write(c.cache); inner != nil {
				dg.Error(inner)
				return
			}
			sucNum.Add(1)
		}(wt)
	}
	return dg.WaitUntilError()
}

func (c *CopyPutStream) Write(p []byte) (n int, err error) {
	length := len(p)
	cur := 0
	for length > 0 {
		next := int(c.rpConfig.BlockSize) - len(c.cache)
		if next > length {
			next = length
		}
		c.cache = append(c.cache, p[cur:cur+next]...)
		if datasize.DataSize(len(c.cache)) == c.rpConfig.BlockSize {
			if err := c.Flush(); err != nil {
				return cur, err
			}
		}
		cur += next
		length -= next
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
	if err := c.Flush(); err != nil {
		return err
	}
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
