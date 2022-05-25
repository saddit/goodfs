package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	global "apiserver/internal/usecase/pool"
)

type RSGetStream struct {
	*rsDecoder
}

type provideStream struct {
	stream io.Reader
	index  int
	err    error
}

func provideGetStream(hash string, locates []string) <-chan *provideStream {
	respChan := make(chan *provideStream, 1)
	go func() {
		defer close(respChan)
		var wg sync.WaitGroup
		for i, ip := range locates {
			wg.Add(1)
			go func(idx int, ip string) {
				defer wg.Done()
				if IsAvailable(ip) {
					reader, e := NewGetStream(ip, fmt.Sprintf("%s.%d", hash, idx))
					respChan <- &provideStream{reader, idx, e}
				} else {
					e := errors.New("data server " + ip + " unavailable")
					respChan <- &provideStream{nil, idx, e}
				}
			}(i, ip)
		}
		wg.Wait()
	}()
	return respChan
}

func NewRSGetStream(size int64, hash string, locates []string) (*RSGetStream, error) {
	readers := make([]io.Reader, global.Config.Rs.AllShards())
	writers := make([]io.Writer, global.Config.Rs.AllShards())
	dsNum := int64(global.Config.Rs.DataShards)
	perSize := (size + dsNum - 1) / dsNum
	ds := GetDataServers()
	var e error
	for r := range provideGetStream(hash, locates) {
		if r.err != nil {
			var ip string
			ds, ip = global.Balancer.Pop(ds)
			writers[r.index], r.err = NewPutStream(ip, fmt.Sprintf("%s.%d", hash, r.index), perSize)
			if r.err != nil {
				return nil, e
			}
			//需更新元数据
			locates[r.index] = ip
			e = ErrNeedUpdateMeta
		} else {
			readers[r.index] = r.stream
		}
	}
	dec := NewDecoder(readers, writers, size)
	return &RSGetStream{dec}, e
}

func (g *RSGetStream) Seek(offset int64, whence int) (int64, error) {
	if whence != io.SeekCurrent {
		panic("only support io.SeekCurrent")
	}
	if offset < 0 {
		return 0, fmt.Errorf("only support forward seek offest")
	}

	//读取offset长度的数据，丢弃于内存
	length := int64(global.Config.Rs.BlockSize())
	buf := bytes.NewBuffer(make([]byte, length))
	for offset > 0 {
		if length > offset {
			//当剩余未读取内容少于BlockSize时 减少读取量
			length = offset
		}
		if _, e := io.CopyN(buf, g, length); e != nil {
			return offset, e
		}
		offset -= length
	}
	buf = nil
	return offset, nil
}

func (g *RSGetStream) Close() error {
	//TODO 用协程优化
	for _, w := range g.writers {
		if w != nil {
			if e := w.(*PutStream).Commit(true); e != nil {
				return e
			}
		}
	}
	return nil
}
