package service

import (
	. "apiserver/internal/usecase/webapi"
	"bytes"
	"common/graceful"
	"common/logs"
	"sync/atomic"
)

//PutStream 需要确保调用了Close或者Commit
//Close() Commit() 可重复调用
type PutStream struct {
	Locate    string
	name      string
	tmpId     string
	compress  bool
	committed *atomic.Value
}

//NewPutStream IO: 发送Post请求到数据服务器
func NewPutStream(ip, name string, size int64, compress bool) (*PutStream, error) {
	id, e := PostTmpObject(ip, name, size)
	if e != nil {
		return nil, e
	}
	flag := &atomic.Value{}
	flag.Store(false)
	res := &PutStream{Locate: ip, name: name, tmpId: id, committed: flag, compress: compress}
	return res, nil
}

//newExistedPutStream 不发送Post请求
func newExistedPutStream(ip, name, id string, compress bool) *PutStream {
	flag := &atomic.Value{}
	flag.Store(false)
	res := &PutStream{Locate: ip, name: name, tmpId: id, committed: flag, compress: compress}
	return res
}

func (p *PutStream) Close() error {
	if p.committed.CompareAndSwap(false, true) {
		return p.Commit(false)
	}
	return nil
}

func (p *PutStream) Write(b []byte) (n int, err error) {
	if err = PatchTmpObject(p.Locate, p.tmpId, bytes.NewBuffer(b)); err != nil {
		return
	}
	return len(b), nil
}

//Commit IO: send commit message and close stream
func (p *PutStream) Commit(ok bool) error {
	if p.committed.CompareAndSwap(false, true) {
		if !ok {
			go func() {
				defer graceful.Recover()
				if err := DeleteTmpObject(p.Locate, p.tmpId); err != nil {
					logs.Std().Error(err)
				}
			}()
			return nil
		}

		return PutTmpObject(p.Locate, p.tmpId, p.compress)
	}
	return nil
}
