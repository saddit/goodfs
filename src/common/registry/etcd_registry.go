package registry

import (
	"bytes"
	. "common/cst"
	"common/logs"
	"common/util"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var registryLog = logs.New("etcd-registry")

type EtcdRegistry struct {
	cli     *clientv3.Client
	cfg     Config
	stdName string
	name    string
	addr    string
}

func NewEtcdRegistry(kv *clientv3.Client, cfg *Config) *EtcdRegistry {
	if cfg.ServerPort == "" {
		panic("registry required ServerPort")
	}
	addr, _ := cfg.RegisterAddr()
	k := fmt.Sprint(cfg.Name, "/", cfg.SID())
	reg := &EtcdRegistry{
		cli:     kv,
		cfg:     *cfg,
		stdName: k,
		name:    k,
		addr:    addr,
	}
	return reg
}

func (e *EtcdRegistry) Key() string {
	return EtcdPrefix.FmtRegistry(e.cfg.Group, e.name)
}

func (e *EtcdRegistry) AsMaster() *EtcdRegistry {
	e.name = fmt.Sprint(e.stdName, "_", "master")
	return e
}

func (e *EtcdRegistry) AsSlave() *EtcdRegistry {
	e.name = fmt.Sprint(e.stdName, "_", "slave")
	return e
}

func (e *EtcdRegistry) GetServiceMapping(name string) map[string]string {
	resp, err := e.cli.Get(context.Background(), EtcdPrefix.FmtRegistry(e.cfg.Group, name), clientv3.WithPrefix())
	if err != nil {
		registryLog.Errorf("get services: %s", err)
		return map[string]string{}
	}
	res := make(map[string]string, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		idx1 := bytes.LastIndexByte(kv.Key, '/')
		if idx1 < 0 {
			continue
		}
		idx2 := bytes.LastIndexByte(kv.Key, '_')
		if idx2 < 0 {
			idx2 = len(kv.Key)
		}
		res[util.BytesToStr(kv.Key[idx1+1:idx2])] = util.BytesToStr(kv.Value)
	}
	return res
}

func (e *EtcdRegistry) GetServices(name string) []string {
	resp, err := e.cli.Get(context.Background(), EtcdPrefix.FmtRegistry(e.cfg.Group, name), clientv3.WithPrefix())
	if err != nil {
		registryLog.Infof("get services: %s", err)
		return []string{}
	}
	res := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		res = append(res, util.BytesToStr(kv.Value))
	}
	return res
}

func (e *EtcdRegistry) GetService(name string, id string) (string, bool) {
	mp := e.GetServiceMapping(name)
	v, ok := mp[id]
	return v, ok
}

func (e *EtcdRegistry) Register(id clientv3.LeaseID) {
	if _, err := e.cli.Put(context.Background(), e.Key(), e.addr, clientv3.WithLease(id)); err != nil {
		registryLog.Errorf("register %s fails: %s", e.addr, err)
		return
	}
	registryLog.Infof("register %s success", e.Key())
}

func (e *EtcdRegistry) Unregister() error {
	_, err := e.cli.Delete(context.Background(), e.Key())
	return err
}
