package service

import (
	"apiserver/config"
	"apiserver/internal/usecase"
	"common/graceful"
	"common/logs"
	"io"
	"sync"

	"github.com/klauspost/reedsolomon"
)

type rsDecoder struct {
	enc          reedsolomon.Encoder
	rsCfg        config.RsConfig
	readers      []io.Reader
	writers      []io.Writer
	cache        []byte
	shardsBuffer []byte
	cursor       int
	total        int64
	size         int64
}

func NewDecoder(readers []io.Reader, writes []io.Writer, size int64, rsCfg *config.RsConfig) *rsDecoder {
	enc, _ := reedsolomon.New(rsCfg.DataShards, rsCfg.ParityShards, reedsolomon.WithAutoGoroutines(rsCfg.BlockPerShard))
	buf := make([]byte, rsCfg.FullSize())
	return &rsDecoder{
		readers:      readers,
		writers:      writes,
		enc:          enc,
		size:         size,
		cache:        buf[:0], // let cache and shardsBuffer using same address to reduce memory allocate and usage
		shardsBuffer: buf,
		rsCfg:        *rsCfg,
	}
}

func (d *rsDecoder) cacheSize() int {
	return len(d.cache) - d.cursor
}

func (d *rsDecoder) Read(bt []byte) (int, error) {
	if d.cacheSize() == 0 {
		// reset cache and cursor
		d.cache = d.cache[:0]
		d.cursor = 0
		// fetch new data
		if e := d.getData(); e != nil {
			return 0, e
		}
	}
	length := len(bt)
	if d.cacheSize() < length {
		length = d.cacheSize()
	}
	copy(bt, d.cache[d.cursor:d.cursor+length])
	d.cursor += length
	return length, nil
}

func (d *rsDecoder) getData() error {
	if d.total == d.size {
		return io.EOF
	}
	if d.total > d.size {
		return usecase.ErrOverRead
	}

	// read shards
	var wg sync.WaitGroup
	shards := make([][]byte, d.rsCfg.AllShards())
	for i := range shards {
		if d.readers[i] == nil {
			continue
		}
		// split shardsBuffer to avoid new memory alloc
		shards[i] = d.shardsBuffer[i*d.rsCfg.BlockPerShard : (i+1)*d.rsCfg.BlockPerShard]
		wg.Add(1)
		go func(idx int) {
			defer graceful.Recover()
			defer wg.Done()
			n, err := io.ReadFull(d.readers[idx], shards[idx])
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				logs.Std().Debugf("read shard %d err: %s", idx, err)
				shards[idx] = nil
			}
			shards[idx] = shards[idx][:n]
		}(i)
	}
	wg.Wait()

	// reconstruct data
	if err := d.enc.Reconstruct(shards); err != nil {
		return err
	}

	// save lost shards
	var err error
	go func() {
		defer graceful.Recover()
		for i, w := range d.writers {
			if w == nil {
				continue
			}
			wg.Add(1)
			go func(idx int, wt io.Writer) {
				defer graceful.Recover()
				defer wg.Done()
				if _, inner := wt.Write(shards[idx]); inner != nil {
					err = inner
					logs.Std().Errorf("rewrite lost shards fail: %s", inner)
					return
				}
			}(i, w)
		}
	}()

	// combine data shards
	for i := range shards[:d.rsCfg.DataShards] {
		shardSize := int64(len(shards[i]))
		// remove padding if exists
		if d.total+shardSize > d.size {
			shardSize = d.size - d.total
		}
		d.cache = append(d.cache, shards[i][:shardSize]...)
		d.total += shardSize
	}

	if !d.rsCfg.RewriteAsync {
		wg.Wait()
		return err
	}

	return nil
}
