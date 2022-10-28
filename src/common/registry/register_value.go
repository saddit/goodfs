package registry

import (
	"fmt"
	"strings"
)

type RegisterValue []byte

func NewRV(httpAddr, rpcAddr string) RegisterValue {
	return RegisterValue(fmt.Sprint(httpAddr, ",", rpcAddr))
}

func (rv RegisterValue) HttpAddr() string {
	v, _ := rv.Addr()
	return v
}

func (rv RegisterValue) RpcAddr() string {
	_, v := rv.Addr()
	return v
}

func (rv RegisterValue) Addr() (httpAddr string, rpcAddr string) {
	sp := strings.Split(string(rv), ",")
	httpAddr = sp[0]
	if len(sp) > 1 {
		rpcAddr = sp[1]
	}
	return
}
