package service

import (
	"apiserver/config"
	"common/util"
	"fmt"
	"io"
)

type RSPutStream struct {
	*rsEncoder
	Locates []string
}

func NewRSPutStream(opt *StreamOption, rsCfg *config.RsConfig) (*RSPutStream, error) {
	if len(opt.Locates) < rsCfg.AllShards() {
		return nil, fmt.Errorf("dataServers ip number mismatch %v", rsCfg.AllShards())
	}
	perShard := rsCfg.ShardSize(opt.Size)
	writers := make([]io.WriteCloser, rsCfg.AllShards())
	wg := util.NewDoneGroup()
	defer wg.Close()
	for i := range writers {
		wg.Todo()
		go func(idx int) {
			defer wg.Done()
			stream, e := NewPutStream(opt.Locates[idx], fmt.Sprintf("%s.%d", opt.Hash, idx), int64(perShard), opt.Compress)
			if e != nil {
				wg.Error(e)
			} else {
				writers[idx] = stream
			}
		}(i)
	}
	if e := wg.WaitUntilError(); e != nil {
		return nil, e
	}
	enc := NewEncoder(writers, rsCfg)
	return &RSPutStream{enc, opt.Locates}, nil
}

func newExistedRSPutStream(ips, ids []string, hash string, compress bool, rsCfg *config.RsConfig) *RSPutStream {
	writers := make([]io.WriteCloser, len(ids))
	for i := range writers {
		writers[i] = newExistedPutStream(ips[i], fmt.Sprintf("%s.%d", hash, i), ids[i], compress)
	}
	return &RSPutStream{NewEncoder(writers, rsCfg), ips}
}

func (p *RSPutStream) Commit(ok bool) error {
	if _, err := p.Flush(); err != nil {
		return err
	}

	wg := util.NewDoneGroup()
	defer wg.Close()
	for _, w := range p.writers {
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
