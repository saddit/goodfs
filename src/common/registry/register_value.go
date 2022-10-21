package registry

import (
	"fmt"
	"strings"
)


type RegisterValue []byte

func NewRV(httpAddr, rpcAddr string) RegisterValue {
	return RegisterValue(fmt.Sprint(httpAddr, rpcAddr))
}

func (rv RegisterValue) HttpAddr() string {
	return strings.Split(string(rv), ",")[0]
}

func (rv RegisterValue) RpcAddr() string {
	return strings.Split(string(rv), ",")[1]
}

func (rv RegisterValue) Addr() (httpAddr string, rpcAddr string) {
	sp := strings.Split(string(rv), ",")
	return sp[0], sp[1]
}