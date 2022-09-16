package logic

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/selector"
	"common/constrant"
	"common/util"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Discovery struct{}

func NewDiscovery() Discovery { return Discovery{} }

func (Discovery) GetDataServers() []string {
	return pool.Discovery.GetServices(pool.Config.Discovery.DataServName)
}

func (Discovery) GetMetaServers(master bool) []string {
	return pool.Discovery.GetServicesWith(pool.Config.Discovery.MetaServName, master)
}

func (Discovery) SelectMetaByGroupID(gid string, defLoc string) string {
	resp, err := pool.Etcd.Get(context.Background(), constrant.EtcdPrefix.FmtPeersInfo(gid, ""), clientv3.WithPrefix())
	if err != nil {
		return defLoc
	}
	if len(resp.Kvs) == 0 {
		return defLoc
	}
	//TODO 缓存group-id的ips
	ips := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var info entity.PeerInfo
		if err = util.DecodeMsgp(&info, kv.Value); err == nil {
			ips = append(ips, fmt.Sprint(info.Location, ":", info.HttpPort))
		}
	}
	if len(ips) == 0 {
		return defLoc
	}
	return selector.NewIPSelector(pool.Balancer, ips).Select()
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
