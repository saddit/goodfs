package service

import (
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/selector"
)

func GetDataServers() []string {
	return pool.Discovery.GetServices(pool.Config.Discovery.DataServName)
}

func GetMetaServers() []string {
	return pool.Discovery.GetServices(pool.Config.Discovery.MetaServName)
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
