package service

import (
	"apiserver/internal/usecase/pool"
	"common/graceful"
	"common/util"
	"fmt"
	"io"
)

type RSPutStream struct {
	*rsEncoder
	Locates []string
}

func NewRSPutStream(ips []string, hash string, size int64) (*RSPutStream, error) {
	rs := pool.Config.Rs
	if len(ips) < rs.AllShards() {
		return nil, fmt.Errorf("dataServers ip number mismatch %v", rs.AllShards())
	}
	ds := int64(rs.DataShards)
	perShard := (size + ds - 1) / ds
	writers := make([]io.WriteCloser, rs.AllShards())
	wg := util.NewDoneGroup()
	for i := range writers {
		wg.Todo()
		go func(idx int) {
			defer graceful.Recover()
			defer wg.Done()
			stream, e := NewPutStream(ips[idx], fmt.Sprintf("%s.%d", hash, idx), perShard)
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
	enc := NewEncoder(writers)
	return &RSPutStream{enc, ips}, nil
}

func newExistedRSPutStream(ips, ids []string, hash string) *RSPutStream {
	writers := make([]io.WriteCloser, len(ids))
	for i := range writers {
		writers[i] = newExistedPutStream(ips[i], fmt.Sprintf("%s.%d", hash, i), ids[i])
	}
	return &RSPutStream{NewEncoder(writers), ips}
}

func (p *RSPutStream) Commit(ok bool) error {
	if _, e := p.Flush(); e != nil {
		return nil
	}

	wg := util.NewDoneGroup()
	defer wg.Close()
	for _, w := range p.writers {
		if util.InstanceOf[Commiter](w) {
			wg.Todo()
			go func(cm Commiter) {
				defer graceful.Recover()
				defer wg.Done()
				if e := cm.Commit(ok); e != nil {
					wg.Error(e)
				}
			}(w.(Commiter))
		}
	}
	return wg.WaitUntilError()
}
