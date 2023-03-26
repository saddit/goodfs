package registry

import (
	"common/cst"
	"common/graceful"
	"common/util"
	"context"
	"strings"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdDiscovery struct {
	cli      *clientv3.Client
	group    string
	services map[string]*serviceList
	context  context.Context
	Close    func()
}

func NewEtcdDiscovery(cli *clientv3.Client, cfg *Config) *EtcdDiscovery {
	hs := make(map[string]*serviceList)
	ctx, cancel := context.WithCancel(context.Background())
	d := &EtcdDiscovery{
		cli:      cli,
		group:    cfg.Group,
		services: hs,
		context:  ctx,
		Close:    cancel,
	}
	for _, s := range cfg.Services {
		d.initService(s)
	}
	return d
}

func (e *EtcdDiscovery) initService(serv string) {
	e.services[serv] = newServiceList()
	go func() {
		defer graceful.Recover()
		// fetch kvs
		prefix := cst.EtcdPrefix.FmtRegistry(e.group, serv)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		res, err := e.cli.Get(ctx, prefix, clientv3.WithPrefix())
		if err != nil {
			registryLog.Warnf("discovery init service %s error: %s", prefix, err)
			return
		}
		// wrap kvs
		mp := make(map[string]string)
		for _, kv := range res.Kvs {
			mp[util.BytesToStr(kv.Key)] = util.BytesToStr(kv.Value)
		}
		// init serv
		e.services[serv].replace(mp)
		// start watch change
		e.asyncWatch(serv, prefix)
	}()
}

func (e *EtcdDiscovery) asyncWatch(serv, prefix string) {
	go func() {
		defer graceful.Recover()
		for {
			var success bool
			ch := e.cli.Watch(e.context, prefix, clientv3.WithPrefix())
			for res := range ch {
				if res.Canceled {
					registryLog.Errorf("dicovery for %s abort: %s", serv, res.Err())
					success = false
					break
				}
				for _, event := range res.Events {
					// Key will be like ${group}/registry/${serv}/${id}_${slave/master}
					key := util.BytesToStr(event.Kv.Key)
					addr := util.BytesToStr(event.Kv.Value)
					switch event.Type {
					case mvccpb.PUT:
						registryLog.Tracef("%s:%s added", serv, addr)
						e.addService(serv, addr, key)
					case mvccpb.DELETE:
						registryLog.Tracef("%s:%s removed", serv, key)
						e.removeService(serv, key)
					}
				}
			}
			// break if canceled by context
			if success {
				break
			}
			// sleep 2 sec before retry
			time.Sleep(2 * time.Second)
		}
	}()
}

func (e *EtcdDiscovery) GetServiceMapping(name string) map[string]string {
	res := make(map[string]string)
	if sl, ok := e.services[name]; ok {
		for k, addr := range sl.copy() {
			idx1 := strings.LastIndexByte(k, '/')
			if idx1 < 0 {
				continue
			}
			idx2 := strings.LastIndexByte(k, '_')
			if idx2 < 0 {
				idx2 = len(k)
			}
			res[k[idx1+1:idx2]] = addr
		}
	}
	return res
}

func (e *EtcdDiscovery) GetServices(name string) []string {
	if sl, ok := e.services[name]; ok {
		return sl.list()
	}
	return []string{}
}

func (e *EtcdDiscovery) GetServiceCount(name string) int {
	if sl, ok := e.services[name]; ok {
		return sl.Len()
	}
	return 0
}

func (e *EtcdDiscovery) GetService(name string, id string) (string, bool) {
	mp := e.GetServiceMapping(name)
	if mp != nil {
		v, ok := mp[id]
		return v, ok
	}
	return "", false
}

func (e *EtcdDiscovery) GetServiceMappingWith(name string, master bool) map[string]string {
	if sl, ok := e.services[name]; ok {
		res := make(map[string]string, len(sl.data))
		for k, addr := range sl.copy() {
			idx := strings.LastIndexByte(k, '/')
			if idx < 0 {
				continue
			}
			sid, role, contains := strings.Cut(k[idx+1:], "_")
			if master {
				if !contains || role == "master" {
					res[sid] = addr
				}
			} else if role == "slave" {
				res[sid] = addr
			}
		}
		return res
	}
	return map[string]string{}
}

func (e *EtcdDiscovery) GetServicesWith(name string, master bool) []string {
	mp := e.GetServiceMappingWith(name, master)
	arr := make([]string, 0, len(mp))
	for _, v := range mp {
		arr = append(arr, v)
	}
	return arr
}

func (e *EtcdDiscovery) addService(name string, value string, key string) {
	e.services[name].add(value, key)
}

func (e *EtcdDiscovery) removeService(name string, key string) {
	e.services[name].remove(key)
}
