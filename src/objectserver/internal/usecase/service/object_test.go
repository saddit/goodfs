package service

import (
	"bytes"
	"path/filepath"

	"testing"
	"github.com/stretchr/testify/assert"
)

func TestWriteReadDeleteFile(t *testing.T) {
	defer func() {
		_ , err := DeleteFile(".", "new_file")
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
	err = GetFile("./new_file", 0, 4096 + 13, buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(buf.String())
	assert.New(t).Equal(4096+13, buf.Len())
}