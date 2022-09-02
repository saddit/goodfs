package registry

import (
	"common/graceful"
	"common/util"
	"context"
	"fmt"
	"strings"

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
				//Key will be like ${serv}_${optional}_${timestamp}
				key := string(event.Kv.Key)
				addr := string(event.Kv.Value)
				switch event.Type {
				case mvccpb.PUT:
					e.addService(serv, addr, key)
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

func (e *EtcdDiscovery) GetServicesWith(name string, master bool) []string {
	s := util.IfElse(master, "master", "slave")

	if sl, ok := e.services[name]; ok {
		arr := make([]string, 0, len(sl.data))
		for k, v := range sl.data {
			if strings.Contains(v, s) {
				arr = append(arr, k)
			}
		}
		return arr
	}
	return []string{}
}

func (e *EtcdDiscovery) addService(name string, addr, key string) {
	e.services[name].add(addr, key)
}

func (e *EtcdDiscovery) removeService(name string, addr string) {
	e.services[name].remove(addr)
}
