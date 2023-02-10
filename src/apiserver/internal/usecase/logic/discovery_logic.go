package logic

import (
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/componet/selector"
	"apiserver/internal/usecase/grpcapi"
	"apiserver/internal/usecase/pool"
)

type Discovery struct{}

func NewDiscovery() Discovery { return Discovery{} }

func (Discovery) GetDataServers() []string {
	return pool.Discovery.GetServices(pool.Config.Discovery.DataServName, false)
}

func (Discovery) GetMetaServerHTTP(id string) string {
	return pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName, false)[id]
}

func (Discovery) GetMetaServerGRPC(id string) string {
	return pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName, true)[id]
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

func (Discovery) SelectMetaServerHttp(metaServerId string) (string, error) {
	metaServs := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName, false)
	ip, ok := metaServs[metaServerId]
	if !ok {
		return "", usecase.ErrServiceUnavailable
	}
	peerIds, _ := grpcapi.GetPeers(ip)
	peerIds = append(peerIds, metaServerId)
	ips := make([]string, 0, len(peerIds))
	for _, id := range peerIds {
		if ip, ok := metaServs[id]; ok {
			ips = append(ips, ip)
		}
	}
	if len(ips) == 0 {
		return "", usecase.ErrServiceUnavailable
	}
	return new(selector.RandomSelector).Select(ips), nil
}

func (Discovery) SelectMetaServerGRPC(metaServerId string) (string, error) {
	metaServs := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName, true)
	ip, ok := metaServs[metaServerId]
	if !ok {
		return "", usecase.ErrServiceUnavailable
	}
	peerIds, _ := grpcapi.GetPeers(ip)
	peerIds = append(peerIds, metaServerId)
	ips := make([]string, 0, len(peerIds))
	for _, id := range peerIds {
		if ip, ok := metaServs[id]; ok {
			ips = append(ips, ip)
		}
	}
	if len(ips) == 0 {
		return "", usecase.ErrServiceUnavailable
	}
	return new(selector.RandomSelector).Select(ips), nil
}

func (Discovery) NewIPSelector(ips []string) selector.IPSelector {
	return selector.IPSelector{Selector: pool.Balancer, IPs: ips}
}

func (d Discovery) NewDataServSelector() selector.IPSelector {
	return d.NewIPSelector(d.GetDataServers())
}
