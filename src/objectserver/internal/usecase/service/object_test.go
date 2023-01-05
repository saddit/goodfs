package service

import (
	"bytes"
	"common/system/disk"
	"os"

	"testing"

	"github.com/ncw/directio"
)

func TestAppendFile(t *testing.T) {
	buf := bytes.NewBufferString("OnlyOneWordInThisFile")
	t.Log("len", buf.Len())
	n, err := AppendFile(".", "new_file", buf)
	t.Log(n, err)
}

func TestGetFile(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 32 * 4096))
	err := GetFile("./new_file", 0, 9999, buf)
	if err != nil {
		t.Fatal(err)
	} 
	t.Log(buf.Len())
	t.Logf("res: '%s'", buf.String())
}

func TestDirectIO(t *testing.T) {
	// starting block
	// block1 := disk.AlignedBlock(2 * directio.BlockSize)
	block1 := make([]byte, 9999)
	for i := 0; i < len(block1); i++ {
		block1[i] = 'F'
	}

	// Write the file
	out, err := disk.OpenFileDirectIO("./new_file", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		t.Fatal("Failed to directio.OpenFile for read", err)
	}
	_, err = disk.AligendWriteTo(out, bytes.NewBuffer(block1), directio.BlockSize)
	if err != nil {
		t.Fatal("Failed to write", err)
	}
	err = out.Close()
	if err != nil {
		t.Fatal("Failed to close writer", err)
	}
}