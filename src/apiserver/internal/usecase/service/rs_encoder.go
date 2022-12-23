package service

import (
	"apiserver/config"
	"github.com/klauspost/reedsolomon"
	"io"
)

type rsEncoder struct {
	writers  []io.WriteCloser
	enc      reedsolomon.Encoder
	cache    []byte
	rsConfig config.RsConfig
}

func NewEncoder(wrs []io.WriteCloser, rsCfg *config.RsConfig) *rsEncoder {
	enc, _ := reedsolomon.New(rsCfg.DataShards, rsCfg.ParityShards)
	return &rsEncoder{
		writers:  wrs,
		enc:      enc,
		cache:    nil,
		rsConfig: *rsCfg,
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
		next := e.rsConfig.BlockSize() - len(e.cache)
		if next > length {
			next = length
		}
		e.cache = append(e.cache, bt[cur:cur+next]...)
		if len(e.cache) == e.rsConfig.BlockSize() {
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
	defer func() { e.cache = make([]byte, 0, e.rsConfig.BlockSize()) }()

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
