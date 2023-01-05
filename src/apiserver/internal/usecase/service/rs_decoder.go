package service

import (
	"apiserver/config"
	"io"

	"github.com/klauspost/reedsolomon"
)

type rsDecoder struct {
	readers   []io.Reader
	writers   []io.Writer
	enc       reedsolomon.Encoder
	size      int64
	cache     []byte
	cacheSize int
	total     int64
	rsCfg     config.RsConfig
}

func NewDecoder(readers []io.Reader, writes []io.Writer, size int64, rsCfg *config.RsConfig) *rsDecoder {
	enc, _ := reedsolomon.New(rsCfg.DataShards, rsCfg.ParityShards)
	return &rsDecoder{
		readers: readers,
		writers: writes,
		enc:     enc,
		size:    size,
		rsCfg:   *rsCfg,
	}
}

func (d *rsDecoder) Read(bt []byte) (int, error) {
	if d.cacheSize == 0 {
		if e := d.getData(); e != nil {
			return 0, e
		}
	}
	length := len(bt)
	if d.cacheSize < length {
		length = d.cacheSize
	}
	d.cacheSize -= length
	copy(bt, d.cache[:length])
	d.cache = d.cache[length:]
	return length, nil
}

func (d *rsDecoder) getData() error {
	if d.total == d.size {
		return io.EOF
	}
	shards := make([][]byte, d.rsCfg.AllShards())
	for i := range shards {
		if d.readers[i] != nil {
			shards[i] = make([]byte, d.rsCfg.BlockPerShard)
			n, e := io.ReadFull(d.readers[i], shards[i])
			if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
				shards[i] = nil
			} else if n != d.rsCfg.BlockPerShard {
				shards[i] = shards[i][:n]
			}
		}
	}
	//缺失修复
	if e := d.enc.ReconstructData(shards); e != nil {
		return e
	}
	for i, w := range d.writers {
		if w != nil {
			if _, e := w.Write(shards[i]); e != nil {
				return nil
			}
		}
	}
	//合并shard
	for i := 0; i < d.rsCfg.DataShards; i++ {
		shardSize := int64(len(shards[i]))
		if d.total+shardSize > d.size {
			shardSize = d.size - d.total
		}
		d.cache = append(d.cache, shards[i][:shardSize]...)
		d.cacheSize += int(shardSize)
		d.total += shardSize
	}
	return nil
}
