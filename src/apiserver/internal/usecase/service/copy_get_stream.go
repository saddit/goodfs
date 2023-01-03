package service

import (
	"apiserver/config"
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/logic"
	"common/graceful"
	"common/util"
	"fmt"
	"io"
)

type CopyGetStream struct {
	reader io.ReadSeekCloser
	writer io.WriteCloser
}

func NewCopyGetStream(hash string, locates []string, rpCfg *config.ReplicationConfig) (*CopyGetStream, error) {
	var getStream io.ReadSeekCloser
	var err error
	var failIds,newLocates []string
	lb := logic.NewDiscovery().NewDataServSelector()
	for idx, loc := range locates {
		id := fmt.Sprint(hash, ".", idx)
		getStream, err = NewGetStream(loc, id)
		if err == nil {
			break
		}
		failIds = append(failIds, id)
		locates[idx] = lb.Select()
		newLocates = append(newLocates, locates[idx])
	}
	if len(failIds) == len(locates) {
		return nil, fmt.Errorf("not found any copies of %s", hash)
	}
	var fixStream io.WriteCloser
	if len(failIds) > rpCfg.ToleranceLossNum() {
		fixStream, err = NewCopyFixStream(failIds, newLocates, rpCfg)
		util.LogErrWithPre("fix copies err", err)
		err = usecase.ErrNeedUpdateMeta
	}
	return &CopyGetStream{
		reader: getStream,
		writer: fixStream,
	}, err
}

func (c *CopyGetStream) Read(p []byte) (n int, err error) {
	if c.writer != nil {
		go func(bt []byte) {
			defer graceful.Recover()
			_, err := c.writer.Write(bt)
			util.LogErrWithPre("copy-get-stream read", err)
		}(p)
	}
	return c.reader.Read(p)
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
