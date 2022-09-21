package util

import (
	"bytes"
	"common/logs"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tinylib/msgp/msgp"
)

const (
	OS_ModeUser = 0700
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

func MD5HashBytes(bt []byte) string {
	crypto := md5.New()
	_, _ = crypto.Write(bt)
	res := crypto.Sum(make([]byte, 0, crypto.BlockSize()))
	return base64.StdEncoding.EncodeToString(res)
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

func InstanceOf[T any](obj any) bool {
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

func IfElse[T any](cond bool, t T, f T) T {
	if cond {
		return t
	}
	return f
}

func GetHost() string {
	var err error
	if ip, err := GetClientIp(); err == nil {
		return ip
	}
	logs.Std().Error(err)
	if host, err := os.Hostname(); err == nil {
		return host
	}
	logs.Std().Error(err)
	return "localhost"
}

func GetHostPort(port string) string {
	return fmt.Sprint(GetHost(), ":", port)
}

func GetHostFromAddr(addr string) string {
	// cut http prefix
	addr = strings.TrimPrefix(addr, "https://")
	addr = strings.TrimPrefix(addr, "http://")
	// if doesn't have port
	if !strings.ContainsRune(addr, ':') {
		return addr
	}
	arr := strings.Split(addr, ":")
	// remove last one which consider to a port
	arr = arr[:len(arr)-1]
	// concat to support ipv6 address
	return strings.Join(arr, "")
}

func GetPort(addr string) string {
	return strings.Split(addr, ":")[1]
}

func GetClientIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", errors.New("can not find the client ip address")
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
func DecodeMsgp[T msgp.Unmarshaler](data T, bt []byte) (err error) {
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