package util

import (
	"bytes"
	"common/logs"
	"encoding/binary"
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

// LookupIP returns ipv4 if success else return empty string.
func LookupIP(addr string) string {
	if ip := ParseIPFromAddr(addr); ip != nil {
		return ip.String()
	}
	return ""
}

// GetHost get host name from environment variables or os.Hostname or GetServerIP
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
	if strings.LastIndexByte(addr, ':') < 0 {
		return addr
	}
	host, _, _ := net.SplitHostPort(addr)
	return host
}

// ParseIPFromAddr Parse to net.IP if found ipv4 else return private or ipv6 or loopback ip. If no found anything, return nil.
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
	var ipv6 net.IP
	for _, ip := range ips {
		if ip.IsLoopback() {
			loopback = ip
		} else if ip.IsPrivate() {
			private = ip
		} else if ip.To4() != nil {
			return ip
		} else if ip.To16() != nil {
			ipv6 = ip
		}
	}

	if private != nil {
		return private
	}

	if ipv6 != nil {
		return ipv6
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

// GetServerIP Return public ipv4 if success else return private or ipv6 or loopback ip. At least, return "127.0.0.1"
func GetServerIP() string {
	if env, ok := os.LookupEnv("SERVER_IP"); ok {
		return env
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		LogErr(err)
		return "127.0.0.1"
	}

	var loopback net.IP
	var private net.IP
	var ipv6 net.IP
	for _, address := range addrs {
		ip := ParseIPFromAddr(address.String())
		if ip != nil {
			if ip.IsPrivate() {
				private = ip
			} else if ip.IsLoopback() {
				loopback = ip
			} else if ip.To4() != nil {
				return ip.String()
			} else if ip.To16() != nil {
				ipv6 = ip
			}
		}
	}

	if private != nil {
		return private.String()
	}

	if ipv6 != nil {
		return ipv6.String()
	}

	return loopback.String()
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
func EncodeMsgp(data msgp.MarshalSizer) ([]byte, error) {
	bt, err := data.MarshalMsg(nil)
	if err != nil {
		return nil, fmt.Errorf("decode-msgp: %w", err)
	}
	return bt, err
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
