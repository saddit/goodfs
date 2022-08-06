package registry

import (
	"common/graceful"
	"context"
	"fmt"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdDiscovery struct {
	*clientv3.Client
	group    string
	services map[string]*serviceList
}

func NewEtcdDiscovery(cli *clientv3.Client, cfg *Config) *EtcdDiscovery {
	m := make(map[string]*serviceList)
	d := &EtcdDiscovery{cli, cfg.Group, m}
	for _, s := range cfg.Services {
		m[s] = newServiceList()
		prefix := fmt.Sprintf("%s/%s", cfg.Group, s)
		ch := d.Watch(context.Background(), prefix, clientv3.WithPrefix())
		d.asyncWatch(s, ch)
	}
	return d
}

func (e *EtcdDiscovery) asyncWatch(serv string, ch clientv3.WatchChan) {
	go func() {
		defer graceful.Recover()
		for res := range ch {
			for _, event := range res.Events {
				//Key will be like ${serv}_${timestamp}
				addr := string(event.Kv.Value)
				switch event.Type {
				case mvccpb.PUT:
					e.addService(serv, addr)
				case mvccpb.DELETE:
					e.removeService(serv, addr)
				}
			}
		}
	}()
}

func (e *EtcdDiscovery) GetServices(name string) []string {
	if sl, ok := e.services[name]; ok {
		return sl.list()
	}
	return []string{}
}

func (e *EtcdDiscovery) addService(name string, addr string) {
	e.services[name].add(addr)
}

func (e *EtcdDiscovery) removeService(name string, addr string) {
	e.services[name].remove(addr)
}
