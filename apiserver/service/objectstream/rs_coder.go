package objectstream

import (
	"github.com/klauspost/reedsolomon"
	"goodfs/apiserver/global"
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

type rsDecoder struct {
	readers   []io.Reader
	writers   []io.Writer
	enc       reedsolomon.Encoder
	size      int64
	cache     []byte
	cacheSize int
	total     int64
}

func NewDecoder(readers []io.Reader, writes []io.Writer, size int64) *rsDecoder {
	enc, _ := reedsolomon.New(global.Config.Rs.DataShards, global.Config.Rs.ParityShards)
	return &rsDecoder{
		readers: readers,
		writers: writes,
		enc:     enc,
		size:    size,
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
	shards := make([][]byte, global.Config.Rs.AllShards())
	for i := range shards {
		if d.readers[i] != nil {
			shards[i] = make([]byte, global.Config.Rs.BlockPerShard)
			n, e := io.ReadFull(d.readers[i], shards[i])
			if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
				shards[i] = nil
			} else if n != global.Config.Rs.BlockPerShard {
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
	for i := 0; i < global.Config.Rs.DataShards; i++ {
		shardSize := int64(len(shards[i]))
		if d.total+shardSize > d.size {
			shardSize -= d.total + shardSize - d.size
		}
		d.cache = append(d.cache, shards[i][:shardSize]...)
		d.cacheSize += int(shardSize)
		d.total += shardSize
	}
	return nil
}
