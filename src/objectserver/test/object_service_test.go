package test

import (
	"bytes"
	"io"
	"objectserver/config"
	global "objectserver/internal/usecase/pool"
	. "objectserver/internal/usecase/service"
	"os"
	"path/filepath"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteReadDeleteFile(t *testing.T) {
	defer func() {
		_, err := DeleteFile(".", "new_file")
		if err != nil {
			t.Error(err)
			return
		}
	}()
	bt := make([]byte, 4096)
	for i := range bt {
		bt[i] = 'A'
	}
	buf := bytes.NewBuffer(bt)
	_, err := WriteFile(filepath.Join(".", "new_file"), buf)
	if err != nil {
		t.Error(err)
		return
	}
	bt = make([]byte, 13)
	for i := range bt {
		bt[i] = 'B'
	}
	buf = bytes.NewBuffer(bt)
	_, err = WriteFile(filepath.Join(".", "new_file"), buf)
	if err != nil {
		t.Error(err)
		return
	}
	buf = bytes.NewBuffer(make([]byte, 0, 32*4096))
	err = GetFile("./new_file", 0, 4096+13, buf)
	if err != nil {
		t.Fatal(err)
	}
	assert.New(t).Equal(4096+13, buf.Len())
}

func TestWriteWithSize(t *testing.T) {
	defer func() {
		_, err := DeleteFile(".", "new_file")
		if err != nil {
			t.Error(err)
			return
		}
	}()
	var bt []byte
	var buffer *bytes.Buffer
	fstSize, secSize := 8000, 4108
	// write first time
	bt = make([]byte, fstSize)
	for i := range bt {
		bt[i] = 'A'
	}
	buffer = bytes.NewBuffer(bt)
	n, err := WriteFileWithSize("./new_file", 0, buffer)
	if err != nil {
		t.Error(err)
		return
	}
	assert.New(t).GreaterOrEqual(n, int64(fstSize))
	// write second time
	bt = make([]byte, secSize)
	for i := range bt {
		bt[i] = 'B'
	}
	buffer = bytes.NewBuffer(bt)
	n, err = WriteFileWithSize("./new_file", int64(fstSize), buffer)
	if err != nil {
		t.Error(err)
		return
	}
	assert.New(t).GreaterOrEqual(n, int64(secSize))
	// read file
	exceptSize := fstSize + secSize - (fstSize+secSize)%4096 + 4096
	fi, err := os.Open("./new_file")
	if err != nil {
		t.Error(err)
		return
	}
	defer fi.Close()
	bt, err = io.ReadAll(fi)
	if err != nil {
		t.Error(err)
		return
	}
	assert.New(t).Equal(exceptSize, len(bt))
	var numOfA, numOfB, numOfOther int
	for _, b := range bt[:fstSize+secSize] {
		if b == 'A' {
			numOfA++
		} else if b == 'B' {
			numOfB++
		} else {
			numOfOther++
		}
	}
	assert.New(t).Equal(0, numOfOther)
	assert.New(t).Equal(fstSize, numOfA)
	assert.New(t).Equal(secSize, numOfB)
}

func TestWriteFileCompress(t *testing.T) {
	defer func() {
		_, err := DeleteFile(".", "new_file")
		if err != nil {
			t.Error(err)
			return
		}
	}()
	var bt []byte
	fstSize, sndSize := 1024*910*3, 1024*1000
	// first write
	bt = make([]byte, fstSize)
	for i := range bt {
		bt[i] = 'A'
	}
	buf := bytes.NewBuffer(bt)
	n, err := WriteFileCompress("./new_file", buf)
	if err != nil {
		t.Fatal(err)
	}
	assert.New(t).Equal(len(bt), int(n))
	// second write
	bt = make([]byte, sndSize)
	for i := range bt {
		bt[i] = 'B'
	}
	buf = bytes.NewBuffer(bt)
	n, err = WriteFileCompress("./new_file", buf)
	if err != nil {
		t.Fatal(err)
	}
	assert.New(t).Equal(len(bt), int(n))
	// read
	buf = bytes.NewBuffer(nil)
	err = GetFileCompress("./new_file", 0, int64(fstSize+sndSize), buf)
	if err != nil {
		t.Fatal(err)
	}
	var numOfA, numOfB, numOfOther int
	for _, b := range buf.Bytes() {
		if b == 'A' {
			numOfA++
		} else if b == 'B' {
			numOfB++
		} else {
			numOfOther++
		}
	}
	assert.New(t).Equal(fstSize, numOfA)
	assert.New(t).Equal(sndSize, numOfB)
	assert.New(t).Equal(0, numOfOther)
}

func TestCommitFile(t *testing.T) {
	defer func() {
		_, err := DeleteFile(".", "new_file")
		if err != nil {
			t.Error(err)
			return
		}
		_, err = DeleteFile(".", "new_file_compress")
		if err != nil {
			t.Error(err)
			return
		}
	}()
	fst, snd := 4096, 3080
	bt := make([]byte, fst)
	for i := range bt {
		bt[i] = 'A'
	}
	buf := bytes.NewBuffer(bt)
	_, err := WriteFile(filepath.Join(".", "new_file"), buf)
	if err != nil {
		t.Error(err)
		return
	}
	bt = make([]byte, snd)
	for i := range bt {
		bt[i] = 'B'
	}
	buf = bytes.NewBuffer(bt)
	_, err = WriteFile(filepath.Join(".", "new_file"), buf)
	if err != nil {
		t.Error(err)
		return
	}
	global.Config = &config.Config{StoragePath: ".", TempPath: "."}
	if err = CommitFile("E:", "new_file", "new_file_compress", true); err != nil {
		t.Error(err)
		return
	}
	buf = bytes.NewBuffer(nil)
	if err = GetFileCompress("./new_file_compress", 0, int64(fst+snd), buf); err != nil {
		t.Error(err)
		return
	}
	var numOfA, numOfB, numOfOther int
	for _, b := range buf.Bytes() {
		if b == 'A' {
			numOfA++
		} else if b == 'B' {
			numOfB++
		} else {
			numOfOther++
		}
	}
	assert.New(t).Equal(fst, numOfA)
	assert.New(t).Equal(snd, numOfB)
	assert.New(t).Equal(0, numOfOther)
}
