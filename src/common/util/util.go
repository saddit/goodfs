package util

import (
	"bytes"
	"common/logs"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
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

// ImmediateTick Immediately tick once then tick interval
// alloc 2 chan, one from time.Tick()
func ImmediateTick(t time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	tk := time.NewTicker(t)
	go func() {
		defer close(ch)
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

// NumToString format number to string by strconv.
// uint and int with base 10.
// float with fmt='f' and prec=10.
// others return empty string
func NumToString(n interface{}) string {
	switch n := n.(type) {
	case int:
		return strconv.Itoa(n)
	case int8:
		return strconv.FormatInt(int64(n), 10)
	case int16:
		return strconv.FormatInt(int64(n), 10)
	case int32:
		return strconv.FormatInt(int64(n), 10)
	case int64:
		return strconv.FormatInt(n, 10)
	case uint:
		return strconv.FormatUint(uint64(n), 10)
	case uint8:
		return strconv.FormatUint(uint64(n), 10)
	case uint16:
		return strconv.FormatUint(uint64(n), 10)
	case uint32:
		return strconv.FormatUint(uint64(n), 10)
	case uint64:
		return strconv.FormatUint(n, 10)
	case float32:
		return strconv.FormatFloat(float64(n), 'f', 10, 32)
	case float64:
		return strconv.FormatFloat(n, 'f', 10, 64)
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

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func IfElse[T any](cond bool, t T, f T) T {
	if cond {
		return t
	}
	return f
}

// LookupIP Return ipv4 if success else return empty string.
func LookupIP(addr string) string {
	if ip := ParseIPFromAddr(addr); ip != nil {
		return ip.String()
	}
	return ""
}

func GetHost() string {
	if host, ok := os.LookupEnv("HOST"); ok {
		return host
	}
	var err error
	if host, err := os.Hostname(); err == nil {
		return host
	}
	logs.Std().Error(err)
	return GetServerIP()
}

func GetHostPort(port string) string {
	return net.JoinHostPort(GetHost(), port)
}

func GetHostFromAddr(addr string) string {
	// cut http prefix
	addr = strings.TrimPrefix(addr, "https://")
	addr = strings.TrimPrefix(addr, "http://")
	host, _, _ := net.SplitHostPort(addr)
	return host
}

// ParseIPFromAddr Parse to net.IP if found ipv4 else return private or loopback ip. If no found anything, return nil.
func ParseIPFromAddr(addr string) net.IP {
	host := GetHostFromAddr(addr)
	if netIP := net.ParseIP(host); netIP != nil {
		return netIP
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		LogErr(err)
		return nil
	}
	var loopback net.IP
	var private net.IP
	for _, ip := range ips {
		if ip.IsLoopback() {
			loopback = ip
		} else if ip.IsPrivate() {
			private = ip
		} else if ip.To4() != nil {
			return ip
		}
	}
	if private != nil {
		return private
	}
	return loopback
}

func GetPortFromAddr(addr string) string {
	// cut http prefix
	addr = strings.TrimPrefix(addr, "https://")
	addr = strings.TrimPrefix(addr, "http://")
	_, port, _ := net.SplitHostPort(addr)
	return port
}

// GetServerIP Return public ipv4 if success else return private or loopback ip. At least, return "127.0.0.1"
func GetServerIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		LogErr(err)
		return "127.0.0.1"
	}

	var loopback net.IP
	var private net.IP
	for _, address := range addrs {
		ip := ParseIPFromAddr(address.String())
		if ip != nil {
			if ip.IsPrivate() {
				private = ip
			} else if ip.IsLoopback() {
				loopback = ip
			} else if ip.To4() != nil {
				return ip.String()
			}
		}
	}

	if private != nil {
		return private.String()
	}

	if loopback != nil {
		return loopback.String()
	}

	return "127.0.0.1"
}

func ToInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		logs.Std().Error(err)
	}
	return i
}

func ToUint64(str string) uint64 {
	i, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		logs.Std().Error(err)
	}
	return i
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
	return
}

// EncodeMsgp encode data with msgp
func EncodeMsgp(data msgp.MarshalSizer) ([]byte, error) {
	return data.MarshalMsg(nil)
}

func MinInt(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func MaxInt(i, j int) int {
	if i < j {
		return j
	}
	return i
}

func MaxUint64(i, j uint64) uint64 {
	if i < j {
		return j
	}
	return i
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
	return *(*[]byte)(unsafe.Pointer(&s))
}
