package util

import (
	"bytes"
	"common/logs"
	xmath "common/util/math"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/tinylib/msgp/msgp"
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

func GobEncode(v any) []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(v); err != nil {
		return nil
	}
	return buf.Bytes()
}

func GobDecode(bt []byte, v any) bool {
	dec := gob.NewDecoder(bytes.NewBuffer(bt))
	if err := dec.Decode(v); err != nil {
		return false
	}
	return true
}

// ImmediateTick Immediately tick once then tick interval
// alloc 2 chan, one from time.Tick()
func ImmediateTick(t time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	tk := time.NewTicker(t)
	go func() {
		defer close(ch)
		defer tk.Stop()
		for t := range tk.C {
			ch <- t
		}
	}()
	ch <- time.Now()
	return ch
}

func InstanceOf[T any](obj any) bool {
	if obj != nil {
		_, ok := obj.(T)
		return ok
	}
	return false
}

func IntString[T xmath.Signed](n T) string {
	return strconv.FormatInt(int64(n), 10)
}

func UIntString[T xmath.Unsigned](n T) string {
	return strconv.FormatUint(uint64(n), 10)
}

func ToString(v any) string {
	if bt, ok := v.([]byte); ok {
		return string(bt)
	}
	return fmt.Sprint(v)
}

// LogErr log if err != nil
func LogErr(err error) {
	if err != nil {
		logs.Std().Error(err)
	}
}

// LogErrWithPre logrus if err != nil
func LogErrWithPre(prefix string, err error) {
	if err != nil {
		logs.Std().Errorf("%s: %v", prefix, err)
	}
}

func CloseAndLog(s io.Closer) {
	if s != nil {
		LogErr(s.Close())
	}
}

// PanicErr panic if err != nil
func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

// IfElse cond ? t : f
func IfElse[T any](cond bool, t T, f T) T {
	if cond {
		return t
	}
	return f
}

func ToInt(str string) int {
	if str == "" {
		return 0
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		logs.Std().Error(err)
	}
	return i
}

func ToInt64(str string) int64 {
	if str == "" {
		return 0
	}
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		logs.Std().Error(err)
	}
	return i
}

func ToUint64(str string) uint64 {
	if str == "" {
		return 0
	}
	i, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		logs.Std().Error(err)
	}
	return i
}

func ToInt32(str string) int32 {
	return int32(ToInt(str))
}

func UnmarshalPtrFromIO[T any](body io.ReadCloser) (*T, error) {
	defer body.Close()
	var data T
	bt, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bt, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func UnmarshalFromIO[T any](body io.ReadCloser) (T, error) {
	defer body.Close()
	var data T
	bt, err := io.ReadAll(body)
	if err != nil {
		return data, err
	}
	if err := json.Unmarshal(bt, &data); err != nil {
		return data, err
	}
	return data, nil
}

// DecodeMsgp decode data by msgp
func DecodeMsgp(data msgp.Unmarshaler, bt []byte) (err error) {
	_, err = data.UnmarshalMsg(bt)
	if err != nil {
		err = fmt.Errorf("encode-msgp: %w", err)
	}
	return
}

// EncodeMsgp encode data with msgp
func EncodeMsgp(data msgp.Marshaler) ([]byte, error) {
	bt, err := data.MarshalMsg(nil)
	if err != nil {
		return nil, fmt.Errorf("decode-msgp: %w", err)
	}
	return bt, err
}

// EncodeArrayMsgp encodes an array of objects by msgp. size of item after encoded must be less than 2 << 32 - 1
func EncodeArrayMsgp[T msgp.Marshaler](arr []T) (res []byte, err error) {
	if len(arr) == 0 {
		return
	}
	buf := make([]byte, 0)
	size := make([]byte, binary.MaxVarintLen32)
	for _, v := range arr {
		buf, err = v.MarshalMsg(buf)
		if err != nil {
			return
		}
		if res == nil {
			res = make([]byte, 0, len(arr)*len(buf))
		}
		if len(buf) > math.MaxInt32 {
			err = fmt.Errorf("msgp encoded object is too large: %d", len(buf))
			return
		}
		binary.PutUvarint(size, uint64(len(buf)))
		res = append(res, size...)
		res = append(res, buf...)
		buf = buf[:0]
	}
	return
}

// DecodeArrayMsgp decodes bytes generated from EncodeArrayMsgp
func DecodeArrayMsgp[T msgp.Unmarshaler](data []byte, constructor func() T) (arr []T, err error) {
	for cur := 0; cur < len(data); {
		item := constructor()
		size, n := binary.Uvarint(data[cur : cur+binary.MaxVarintLen32])
		if n <= 0 {
			err = errors.New("decode array msgp: format err")
			return
		}
		cur += binary.MaxVarintLen32
		if _, err = item.UnmarshalMsg(data[cur : cur+int(size)]); err != nil {
			return
		}
		cur += int(size)
		arr = append(arr, item)
	}
	return
}

func PagingOffset(page, size, total int) (int, int, bool) {
	if page == 0 {
		return 0, 0, false
	}
	offset := (page - 1) * size
	if offset >= total {
		return 0, 0, false
	}
	end := offset + size
	if end > total {
		end = total
	}
	return offset, end, true
}

// BytesToStr performs unholy acts to avoid allocations
func BytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StrToBytes performs unholy acts to avoid allocations
func StrToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func IntToBytes(i uint64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, i)
	return buf
}

func BytesToInt(b []byte) uint64 {
	i, _ := binary.ReadUvarint(bytes.NewBuffer(b))
	return i
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	rd := rand.New(rand.NewSource(time.Now().UnixMilli()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rd.Intn(len(letterRunes))]
	}
	return string(b)
}
