package objectstream

import (
	"errors"
	"fmt"
	"goodfs/apiserver/global"
	"goodfs/apiserver/model/meta"
	"goodfs/apiserver/service/dataserv"
	"io"
)

var (
	ErrNeedUpdateMeta = errors.New("metadata has changed unavailable server's location")
)

type RSPutStream struct {
	*rsEncoder
	Locates []string
}

func NewRSPutStream(ips []string, hash string, size int64) (*RSPutStream, error) {
	rs := global.Config.Rs
	if len(ips) < rs.AllShards() {
		return nil, fmt.Errorf("dataServers ip number mismatch %v", rs.AllShards())
	}
	ds := int64(rs.DataShards)
	perShard := (size + ds - 1) / ds
	writers := make([]io.Writer, rs.AllShards())
	var e error
	for i := range writers {
		writers[i], e = NewPutStream(ips[i], fmt.Sprintf("%s.%d", hash, i), perShard)
		if e != nil {
			return nil, e
		}
	}
	enc := NewEncoder(writers)
	return &RSPutStream{enc, ips}, nil
}

func (p *RSPutStream) Commit(ok bool) error {
	var e error
	if _, e = p.Flush(); e != nil {
		return nil
	}

	for _, w := range p.writers {
		if e = w.(*PutStream).Commit(ok); e != nil {
			return e
		}
	}
	return nil
}

type RSGetStream struct {
	*rsDecoder
}

func NewRSGetStream(ver *meta.MetaVersion) (*RSGetStream, error) {
	ds := dataserv.GetDataServers()
	readers := make([]io.Reader, global.Config.Rs.AllShards())
	writers := make([]io.Writer, global.Config.Rs.AllShards())
	dsNum := int64(global.Config.Rs.DataShards)
	perSize := (ver.Size + dsNum - 1) / dsNum
	var e error
	for i, ip := range ver.Locate {
		if dataserv.IsAvailable(ip) {
			readers[i], e = NewGetStream(ip, fmt.Sprintf("%s.%d", ver.Hash, i))
			if e != nil {
				return nil, e
			}
		} else {
			ds, ip = global.Balancer.Pop(ds)
			writers[i], e = NewPutStream(ip, fmt.Sprintf("%s.%d", ver.Hash, i), perSize)
			if e != nil {
				return nil, e
			}
			//需更新元数据
			ver.Locate[i] = ip
			e = ErrNeedUpdateMeta
		}
	}
	dec := NewDecoder(readers, writers, ver.Size)
	return &RSGetStream{dec}, e
}

func (g *RSGetStream) Close() error {
	for _, w := range g.writers {
		if w != nil {
			if e := w.(*PutStream).Commit(true); e != nil {
				return e
			}
		}
	}
	return nil
}
