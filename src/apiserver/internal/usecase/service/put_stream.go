package service

import (
	. "apiserver/internal/usecase/webapi"
	"bytes"
	"common/graceful"
	"sync/atomic"
)

//PutStream 需要确保调用了Close或者Commit
//Close() Commit() 可重复调用
type PutStream struct {
	Locate    string
	name      string
	tmpId     string
	committed *atomic.Value
}

//NewPutStream IO: 发送Post请求到数据服务器
func NewPutStream(ip, name string, size int64) (*PutStream, error) {
	id, e := PostTmpObject(ip, name, size)
	if e != nil {
		return nil, e
	}
	flag := &atomic.Value{}
	flag.Store(false)
	res := &PutStream{Locate: ip, name: name, tmpId: id, committed: flag}
	return res, nil
}

//newExistedPutStream 不发送Post请求
func newExistedPutStream(ip, name, id string) *PutStream {
	flag := &atomic.Value{}
	flag.Store(false)
	res := &PutStream{Locate: ip, name: name, tmpId: id, committed: flag}
	return res
}

func (p *PutStream) Close() error {
	if p.committed.CompareAndSwap(false, true) {
		return p.Commit(false)
	}
	return nil
}

func (p *PutStream) Write(b []byte) (n int, err error) {
	if err := PatchTmpObject(p.Locate, p.tmpId, bytes.NewBuffer(b)); err != nil {
		return 0, nil
	}
	return len(b), nil
}

//Commit IO: send commit message and close stream
func (p *PutStream) Commit(ok bool) error {
	if p.committed.CompareAndSwap(false, true) {
		if !ok {
			go func() {
				defer graceful.Recover()
				DeleteTmpObject(p.Locate, p.tmpId)
			}()
			return nil
		}

		return PutTmpObject(p.Locate, p.tmpId, p.name)
	}
	return nil
}
