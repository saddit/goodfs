package service

import (
	"apiserver/internal/usecase/pool"
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
	var e error
	//TODO 用协程优化
	for i := range writers {
		writers[i], e = NewPutStream(ips[i], fmt.Sprintf("%s.%d", hash, i), perShard)
	}
	if e != nil {
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
	var e error
	if _, e = p.Flush(); e != nil {
		return nil
	}

	//TODO 用协程优化
	for _, w := range p.writers {
		if e = w.(*PutStream).Commit(ok); e != nil {
			return e
		}
	}
	return nil
}
