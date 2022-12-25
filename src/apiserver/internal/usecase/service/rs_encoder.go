package service

import (
	"apiserver/config"
	"common/util"
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
			if err := e.Flush(); err != nil {
				return cur, err
			}
		}
		cur += next
		length -= next
	}
	return len(bt), nil
}

func (e *rsEncoder) Flush() error {
	if len(e.cache) == 0 {
		return nil
	}
	defer func() { e.cache = make([]byte, 0, e.rsConfig.BlockSize()) }()

	shards, err := e.enc.Split(e.cache)
	if err != nil {
		return err
	}
	err = e.enc.Encode(shards)
	if err != nil {
		return err
	}

	dg := util.NewDoneGroup()
	defer dg.Close()
	for i, v := range shards {
		dg.Todo()
		go func() {
			defer dg.Done()
			if _, err := e.writers[i].Write(v); err != nil {
				dg.Error(err)
				return
			}
		}()
	}
	return dg.WaitUntilError()
}
