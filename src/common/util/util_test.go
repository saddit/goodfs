package util_test

import (
	"common/proto/msg"
	"common/util"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestEncodeDecodeArrayByMsgpack(t *testing.T) {
	var arr []*msg.Metadata
	for i := 0; i < 10; i++ {
		item := &msg.Metadata{Name: fmt.Sprint(i)}
		if i%2 == 0 {
			item.Bucket = "%2F/"
			item.Extra = &msg.Extra{LastVersion: i + 100}
		} else {
			item.Bucket = util.RandString(i + 10)
			item.Extra = &msg.Extra{Total: i + rand.Int()}
		}
		arr = append(arr, item)
	}
	bt, err := util.EncodeArrayMsgp(arr)
	if err != nil {
		t.Fatal(err)
	}
	res, err := util.DecodeArrayMsgp(bt, func() *msg.Metadata { return &msg.Metadata{} })
	if err != nil {
		t.Fatal(err)
	}
	ast := assert.New(t)
	ast.Equal(len(arr), len(res))
	for i, v := range res {
		ast.Equal(v.Name, arr[i].Name)
		ast.Equal(v.Extra.LastVersion, arr[i].Extra.LastVersion)
		ast.Equal(v.Bucket, arr[i].Bucket)
	}
}

func BenchmarkDecodeArrayMsgpOf50Items(b *testing.B) {
	var arr []*msg.Metadata
	for i := 0; i < 50; i++ {
		item := &msg.Metadata{Name: util.RandString(i + 10)}
		item.Bucket = util.RandString(i + 10)
		item.Extra = &msg.Extra{Total: i + rand.Int()}
		arr = append(arr, item)
	}
	bt, _ := util.EncodeArrayMsgp(arr)
	for i := 0; i < b.N; i++ {
		_, _ = util.DecodeArrayMsgp(bt, func() *msg.Metadata { return new(msg.Metadata) })
	}
}

func BenchmarkEncodeArrayMsgpOf50Items(b *testing.B) {
	var arr []*msg.Metadata
	for i := 0; i < 50; i++ {
		item := &msg.Metadata{Name: util.RandString(i + 10)}
		item.Bucket = util.RandString(i + 10)
		item.Extra = &msg.Extra{Total: i + rand.Int()}
		arr = append(arr, item)
	}
	for i := 0; i < b.N; i++ {
		_, _ = util.EncodeArrayMsgp(arr)
	}
}

func BenchmarkDecodeArrayMsgpOf15Items(b *testing.B) {
	var arr []*msg.Metadata
	for i := 0; i < 15; i++ {
		item := &msg.Metadata{Name: util.RandString(i + 10)}
		item.Bucket = util.RandString(i + 10)
		item.Extra = &msg.Extra{Total: i + rand.Int()}
		arr = append(arr, item)
	}
	bt, _ := util.EncodeArrayMsgp(arr)
	for i := 0; i < b.N; i++ {
		_, _ = util.DecodeArrayMsgp(bt, func() *msg.Metadata { return new(msg.Metadata) })
	}
}

func BenchmarkEncodeArrayMsgpOf15Items(b *testing.B) {
	var arr []*msg.Metadata
	for i := 0; i < 15; i++ {
		item := &msg.Metadata{Name: util.RandString(i + 10)}
		item.Bucket = util.RandString(i + 10)
		item.Extra = &msg.Extra{Total: i + rand.Int()}
		arr = append(arr, item)
	}
	for i := 0; i < b.N; i++ {
		_, _ = util.EncodeArrayMsgp(arr)
	}
}
