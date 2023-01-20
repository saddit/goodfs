package service

import (
	"apiserver/config"
	"common/logs"
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
			n, err := io.ReadFull(d.readers[i], shards[i])
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				logs.Std().Debugf("read shard %d err: %s", i, err)
				shards[i] = nil
				continue
			}
			shards[i] = shards[i][:n]
		}
	}
	// reconstruct all if parity shards lost
	reconstructFunc := d.enc.ReconstructData
	for i := d.rsCfg.DataShards; i < d.rsCfg.AllShards(); i++ {
		if d.writers[i] != nil {
			reconstructFunc = d.enc.Reconstruct
			break
		}
	}
	if err := reconstructFunc(shards); err != nil {
		return err
	}
	for i, w := range d.writers {
		if w != nil {
			if _, err := w.Write(shards[i]); err != nil {
				return err
			}
		}
	}
	// combine data shards
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
