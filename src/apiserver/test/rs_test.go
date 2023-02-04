package test

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/reedsolomon"
)

func BenchmarkReedSolomonEncode(b *testing.B) {
	data := make([]byte, 10<<20)
	for i := range data {
		data[i] = byte(i % 128)
	}
	buf := make([]byte, 8192*4)
	encoder, err := reedsolomon.New(4, 2)
	if err != nil {
		b.Error(err)
		return
	}
	for i := 0; i < b.N; i++ {
		dataIO := bytes.NewBuffer(data)
		encShards := make([][]byte, 4+2)
		for {
			n, err := io.ReadFull(dataIO, buf)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			if err != nil {
				b.Error(err)
				return
			}
			shards, err := encoder.Split(buf[:n])
			if err != nil {
				b.Error(err)
				return
			}
			if err = encoder.Encode(shards); err != nil {
				b.Error(err)
				return
			}
			for i := range shards {
				encShards[i] = append(encShards[i], shards[i]...)
			}
		}
	}
}

func BenchmarkReedSolomonEncodeWithAutoGoroutine(b *testing.B) {
	data := make([]byte, 10<<20)
	for i := range data {
		data[i] = byte(i % 128)
	}
	buf := make([]byte, 8192*4)
	// WithAutoGoroutines imporve 7.68% speed and reduce 25.45% allocs
	encoder, err := reedsolomon.New(4, 2, reedsolomon.WithAutoGoroutines(8192))
	if err != nil {
		b.Error(err)
		return
	}
	for i := 0; i < b.N; i++ {
		dataIO := bytes.NewBuffer(data)
		encShards := make([][]byte, 4+2)
		for {
			n, err := io.ReadFull(dataIO, buf)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			if err != nil {
				b.Error(err)
				return
			}
			shards, err := encoder.Split(buf[:n])
			if err != nil {
				b.Error(err)
				return
			}
			if err = encoder.Encode(shards); err != nil {
				b.Error(err)
				return
			}
			for i := range shards {
				encShards[i] = append(encShards[i], shards[i]...)
			}
		}
	}
}

func BenchmarkReedSolomonStreamEncode(b *testing.B) {
	data := make([]byte, 10<<20)
	for i := range data {
		data[i] = byte(i % 128)
	}
	encoder, err := reedsolomon.NewStreamC(4, 2, true, true)
	if err != nil {
		b.Error(err)
		return
	}
	var shardsRd = make([]io.Reader, 4)
	var shardsWt = make([]io.Writer, 4)
	var encShards = make([]io.Writer, 2)
	for i := 0; i < b.N; i++ {
		dataIO := bytes.NewBuffer(data)
		for i := 0; i < 4; i++ {
			bf := bytes.NewBuffer(nil)
			shardsRd[i] = bf
			shardsWt[i] = bf
			if i < 2 {
				encShards[i] = bytes.NewBuffer(nil)
			}
		}
		if err := encoder.Split(dataIO, shardsWt, 10<<20); err != nil {
			b.Error(err)
			return
		}
		if err := encoder.Encode(shardsRd, encShards); err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkReedSolomonStreamEncodeWithAutoGoroutine(b *testing.B) {
	data := make([]byte, 10<<20)
	for i := range data {
		data[i] = byte(i % 128)
	}
	// WithAutoGoroutines imporve 5.64% speed and reduce 1.59% allocs
	encoder, err := reedsolomon.NewStreamC(4, 2, true, true, reedsolomon.WithAutoGoroutines((10<<20)/4))
	if err != nil {
		b.Error(err)
		return
	}
	var shardsRd = make([]io.Reader, 4)
	var shardsWt = make([]io.Writer, 4)
	var encShards = make([]io.Writer, 2)
	for i := 0; i < b.N; i++ {
		dataIO := bytes.NewBuffer(data)
		for i := 0; i < 4; i++ {
			bf := bytes.NewBuffer(nil)
			shardsRd[i] = bf
			shardsWt[i] = bf
			if i < 2 {
				encShards[i] = bytes.NewBuffer(nil)
			}
		}
		if err := encoder.Split(dataIO, shardsWt, 10<<20); err != nil {
			b.Error(err)
			return
		}
		if err := encoder.Encode(shardsRd, encShards); err != nil {
			b.Error(err)
			return
		}
	}
}

func reedSolomonEncodeSave() error {
	data := make([]byte, 10<<20)
	for i := range data {
		data[i] = byte(i % 128)
	}
	encoder, err := reedsolomon.New(4, 2)
	if err != nil {
		return err
	}
	shards, err := encoder.Split(data)
	if err != nil {
		return err
	}
	if err = encoder.Encode(shards); err != nil {
		return err
	}
	for i := range shards[:4] {
		_ = os.WriteFile(fmt.Sprint("./temp/shard.data.", i), shards[i], os.ModePerm)
	}
	for i := range shards[4:] {
		_ = os.WriteFile(fmt.Sprint("./temp/shard.parity.", i), shards[i], os.ModePerm)
	}
	return nil
}

func BenchmarkReedSolomonDecode(b *testing.B) {
	_ = reedSolomonEncodeSave()
	dataShards := make([][]byte, 0, 6)

	_ = filepath.Walk("./temp", func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		bt, _ := os.ReadFile(path)
		dataShards = append(dataShards, bt)
		return nil
	})

	enc, _ := reedsolomon.New(4, 2)

	for i := 0; i < b.N; i++ {
		dataShards[0] = nil
		dataShards[2] = nil
		if err := enc.ReconstructData(dataShards); err != nil {
			b.Fatal(err)
		}
	
		if err := enc.Join(bytes.NewBuffer(nil), dataShards, 10 << 20); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReedSolomonDecodeWithAutoGoroutine(b *testing.B) {
	_ = reedSolomonEncodeSave()
	dataShards := make([][]byte, 0, 6)

	_ = filepath.Walk("./temp", func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		bt, _ := os.ReadFile(path)
		dataShards = append(dataShards, bt)
		return nil
	})

	enc, _ := reedsolomon.New(4, 2, reedsolomon.WithAutoGoroutines((10 << 20) / 4))

	for i := 0; i < b.N; i++ {
		dataShards[0] = nil
		dataShards[2] = nil
		if err := enc.ReconstructData(dataShards); err != nil {
			b.Fatal(err)
		}
	
		if err := enc.Join(bytes.NewBuffer(nil), dataShards, 10 << 20); err != nil {
			b.Fatal(err)
		}
	}
}
