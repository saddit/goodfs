package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func GetFileExt(fileName string, withDot bool) (string, bool) {
	r := GetFileExtOrDefault(fileName, withDot, "")
	return r, r != ""
}

func GetFileExtOrDefault(fileName string, withDot bool, def string) string {
	idx := strings.LastIndex(fileName, ".")
	if idx == -1 {
		return def
	}
	if !withDot {
		idx++
	}
	return fileName[idx:]
}

//SHA256Hash sha256算法对二进制流进行计算
func SHA256Hash(reader io.Reader) string {
	crypto := sha256.New()
	if _, e := io.CopyBuffer(crypto, reader, make([]byte, 2048)); e == nil {
		b := crypto.Sum(make([]byte, 0, crypto.Size()))
		return base64.StdEncoding.EncodeToString(b)
	}
	return ""
}

func GobEncode(v interface{}) []byte {
	// encode
	buf := new(bytes.Buffer)   // 创建一个buffer区
	enc := gob.NewEncoder(buf) // 创建新的需要转化二进制区域对象
	// 将数据转化为二进制流
	if err := enc.Encode(v); err != nil {
		return nil
	}
	return buf.Bytes()
}

func GobDecodeGen[T interface{}](bt []byte) (*T, bool) {
	var res T
	dec := gob.NewDecoder(bytes.NewBuffer(bt)) // 创建一个对象 把需要转化的对象放入
	// 进行流转化
	if err := dec.Decode(&res); err != nil {
		return nil, false
	}
	return &res, true
}

func GobDecodeGen2[T interface{}](bt []byte, v *T) bool {
	var res T
	dec := gob.NewDecoder(bytes.NewBuffer(bt)) // 创建一个对象 把需要转化的对象放入
	// 进行流转化
	if err := dec.Decode(&res); err != nil {
		return false
	}
	*v = res
	return true
}

//ImmediateTick Immediately tick once then tick interval
//alloc 2 chan, one from time.Tick()
func ImmediateTick(t time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	tk := time.Tick(t)
	go func() {
		defer close(ch)
		for t := range tk {
			ch <- t
		}
	}()
	ch <- time.Now()
	return ch
}

func InstanceOf[T interface{}](obj interface{}) bool {
	if obj != nil {
		_, ok := obj.(T)
		return ok
	}
	return false
}

//NumToString format number to string by strconv.
//uint and int with base 10.
//float with fmt='f' and prec=10.
//others return empty string
func NumToString(n interface{}) string {
	switch n.(type) {
	case int:
		return strconv.Itoa(n.(int))
	case int8:
		return strconv.FormatInt(int64(n.(int8)), 10)
	case int16:
		return strconv.FormatInt(int64(n.(int16)), 10)
	case int32:
		return strconv.FormatInt(int64(n.(int32)), 10)
	case int64:
		return strconv.FormatInt(n.(int64), 10)
	case uint:
		return strconv.FormatUint(uint64(n.(uint)), 10)
	case uint8:
		return strconv.FormatUint(uint64(n.(uint8)), 10)
	case uint16:
		return strconv.FormatUint(uint64(n.(uint16)), 10)
	case uint32:
		return strconv.FormatUint(uint64(n.(uint32)), 10)
	case uint64:
		return strconv.FormatUint(n.(uint64), 10)
	case float32:
		return strconv.FormatFloat(float64(n.(float32)), 'f', 10, 32)
	case float64:
		return strconv.FormatFloat(n.(float64), 'f', 10, 64)
	default:
		return ""
	}
}

func ToString(v any) string {
	return fmt.Sprint(v)
}

// LogErr logrus if err != nil
func LogErr(err error) {
	if err != nil {
		logrus.Error(err)
	}
}
