package service

import (
	"apiserver/internal/usecase/selector"
)

func GetDataServers() []string {
	//TODO 从注册中心获取
	ds := make([]string, 0)
	return ds
}

func SelectDataServer(sel selector.Selector, size int) []string {
	ds := GetDataServers()
	if len(ds) == 0 {
		return []string{}
	}
	serv := make([]string, size)
	lb := selector.IPSelector{Selector: sel, IPs: ds}
	for i := 0; i < size; i++ {
		serv[i] = lb.Select()
	}
	return serv
}
