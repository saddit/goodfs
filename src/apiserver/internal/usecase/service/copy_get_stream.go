package service

import (
	"apiserver/config"
	"apiserver/internal/usecase/logic"
	"common/util"
	"fmt"
	"io"
)

type CopyGetStream struct {
	reader io.ReadSeekCloser
	writer io.WriteCloser
}

func NewCopyGetStream(opt *StreamOption, rpCfg *config.ReplicationConfig) (*CopyGetStream, error) {
	var getStream io.ReadSeekCloser
	var err error
	var failIds, newLocates []string
	lb := logic.NewDiscovery().NewDataServSelector()
	for idx, loc := range opt.Locates {
		id := fmt.Sprint(opt.Hash, ".", idx)
		getStream, err = NewGetStream(loc, id, opt.Size, opt.Compress)
		if err == nil {
			break
		}
		failIds = append(failIds, id)
		opt.Locates[idx] = lb.Select()
		newLocates = append(newLocates, opt.Locates[idx])
	}
	if len(failIds) == len(opt.Locates) {
		return nil, fmt.Errorf("not found any copies of %s", opt.Hash)
	}
	var fixStream io.WriteCloser
	if len(failIds) > rpCfg.ToleranceLossNum() {
		fixStream, err = NewCopyFixStream(failIds, newLocates, opt, rpCfg)
		util.LogErrWithPre("fix copies err", err)
	}
	return &CopyGetStream{
		reader: getStream,
		writer: fixStream,
	}, nil
}

func (c *CopyGetStream) Read(p []byte) (n int, err error) {
	n, err = c.reader.Read(p)
	if c.writer != nil && n > 0 {
		if n, err = c.writer.Write(p[:n]); err != nil {
			return
		}
	}
	return
}

func (c *CopyGetStream) Seek(offset int64, whence int) (int64, error) {
	return c.reader.Seek(offset, whence)
}

func (c *CopyGetStream) Close() (err error) {
	if c.writer != nil {
		if inner := c.writer.Close(); inner != nil {
			err = inner
		}
	}
	if inner := c.reader.Close(); inner != nil {
		err = inner
	}
	return
}
