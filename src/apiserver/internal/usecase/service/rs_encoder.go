package service

import (
	global "apiserver/internal/usecase/pool"
	"github.com/klauspost/reedsolomon"
	"io"
)

type rsEncoder struct {
	writers []io.WriteCloser
	enc     reedsolomon.Encoder
	cache   []byte
}

func NewEncoder(wrs []io.WriteCloser) *rsEncoder {
	enc, _ := reedsolomon.New(global.Config.Rs.DataShards, global.Config.Rs.ParityShards)
	return &rsEncoder{
		writers: wrs,
		enc:     enc,
		cache:   nil,
	}
}

func (e *rsEncoder) Close() error {
	var err error
	for _, w := range e.writers {
		if e := w.Close(); e != nil {
			err = e
		}
	}
	return err
}

func (e *rsEncoder) Write(bt []byte) (int, error) {
	length := len(bt)
	cur := 0
	for length != 0 {
		next := global.Config.Rs.BlockSize() - len(e.cache)
		if next > length {
			next = length
		}
		e.cache = append(e.cache, bt[cur:cur+next]...)
		if len(e.cache) == global.Config.Rs.BlockSize() {
			i, err := e.Flush()
			if err != nil {
				return i, err
			}
		}
		cur += next
		length -= next
	}
	return len(bt), nil
}

func (e *rsEncoder) Flush() (int, error) {
	if len(e.cache) == 0 {
		return 0, nil
	}
	defer func() { e.cache = make([]byte, 0, global.Config.Rs.BlockSize()) }()

	shards, err := e.enc.Split(e.cache)
	if err != nil {
		return 0, err
	}
	err = e.enc.Encode(shards)
	if err != nil {
		return 0, err
	}

	l := 0
	for i, v := range shards {
		j, err := e.writers[i].Write(v)
		l += j
		if err != nil {
			return l, err
		}
	}
	return l, nil
}
