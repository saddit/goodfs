package logic

import (
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/selector"
)

type Discovery struct{}

func NewDiscovery() Discovery {return Discovery{}}

func (Discovery) GetDataServers() []string {
	return pool.Discovery.GetServices(pool.Config.Discovery.DataServName)
}

func (Discovery) GetMetaServers(master bool) []string {
	return pool.Discovery.GetServicesWith(pool.Config.Discovery.MetaServName, master)
}

func (d Discovery) SelectDataServer(sel selector.Selector, size int) []string {
	ds := d.GetDataServers()
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