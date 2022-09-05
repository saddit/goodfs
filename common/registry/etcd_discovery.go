package registry

import (
	"common/graceful"
	"common/util"
	"context"
	"fmt"
	"strings"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdDiscovery struct {
	cli *clientv3.Client
	group    string
	services map[string]*serviceList
}

func NewEtcdDiscovery(cli *clientv3.Client, cfg *Config) *EtcdDiscovery {
	m := make(map[string]*serviceList)
	d := &EtcdDiscovery{cli, cfg.Group, m}
	for _, s := range cfg.Services {
		d.initService(s)
	}
	return d
}

func (e *EtcdDiscovery) initService(serv string) {
	// watch kv changing
	e.services[serv] = newServiceList()
	prefix := fmt.Sprintf("%s/%s", e.group, serv)
	ch := e.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	e.asyncWatch(serv, ch)
	// get original kvs
	go func() {
		defer graceful.Recover()
		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()
		res, err := e.cli.Get(ctx, prefix, clientv3.WithPrefix())
		if err != nil {
			log.Warnf("discovery init service %s error: %s", prefix, err)
			return
		}
		for _, kv := range res.Kvs {
			e.services[serv].add(string(kv.Value), string(kv.Key))
		}
	}()
}

func (e *EtcdDiscovery) asyncWatch(serv string, ch clientv3.WatchChan) {
	go func() {
		defer graceful.Recover()
		for res := range ch {
			for _, event := range res.Events {
				//Key will be like ${serv}_${timestamp}_${optional}
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
			if strings.HasSuffix(v, s) {
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
