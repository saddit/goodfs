package service

import (
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/webapi"
	"bytes"
	"common/graceful"
	"common/logs"
	"sync/atomic"
)

// PutStream transaction put stream
type PutStream struct {
	Locate    string
	name      string
	tmpId     string
	compress  bool
	committed *atomic.Bool
}

// NewPutStream IO: sending POST request to server
func NewPutStream(ip, name string, size int64, compress bool) (*PutStream, error) {
	id, e := webapi.PostTmpObject(ip, name, size)
	if e != nil {
		return nil, e
	}
	res := &PutStream{Locate: ip, name: name, tmpId: id, committed: &atomic.Bool{}, compress: compress}
	return res, nil
}

// newExistedPutStream skip POST request to continue a transfer
func newExistedPutStream(ip, name, id string, compress bool) *PutStream {
	res := &PutStream{Locate: ip, name: name, tmpId: id, committed: &atomic.Bool{}, compress: compress}
	return res
}

func (p *PutStream) Close() error {
	p.committed.CompareAndSwap(false, true)
	return nil
}

func (p *PutStream) Write(b []byte) (n int, err error) {
	if p.committed.Load() {
		return 0, usecase.ErrStreamClosed
	}
	if err = webapi.PatchTmpObject(p.Locate, p.tmpId, bytes.NewBuffer(b)); err != nil {
		return
	}
	return len(b), nil
}

// Commit IO: send commit message and close stream
func (p *PutStream) Commit(ok bool) error {
	if p.committed.CompareAndSwap(false, true) {
		if !ok {
			go func() {
				defer graceful.Recover()
				if err := webapi.DeleteTmpObject(p.Locate, p.tmpId); err != nil {
					logs.Std().Error(err)
				}
			}()
			return nil
		}

		return webapi.PutTmpObject(p.Locate, p.tmpId, p.compress)
	}
	return nil
}
