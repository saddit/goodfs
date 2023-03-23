package registry

import (
	"bytes"
	. "common/cst"
	"common/graceful"
	"common/logs"
	"common/util"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var registryLog = logs.New("etcd-registry")

type EtcdRegistry struct {
	cli         *clientv3.Client
	cfg         Config
	leaseId     clientv3.LeaseID
	leaseIdChan chan clientv3.LeaseID
	stdName     string
	name        string
	addr        string
	stopFn      func()
}

func NewEtcdRegistry(kv *clientv3.Client, cfg *Config) *EtcdRegistry {
	if cfg.ServerPort == "" {
		panic("registry required ServerPort")
	}
	addr, _ := cfg.RegisterAddr()
	k := fmt.Sprint(cfg.Name, "/", cfg.SID())
	return &EtcdRegistry{
		cli:         kv,
		cfg:         *cfg,
		leaseId:     -1,
		leaseIdChan: make(chan clientv3.LeaseID, 1),
		stdName:     k,
		name:        k,
		addr:        addr,
		stopFn:      func() {},
	}
}

func (e *EtcdRegistry) Key() string {
	return EtcdPrefix.FmtRegistry(e.cfg.Group, e.name)
}

func (e *EtcdRegistry) AsMaster() *EtcdRegistry {
	// metaserver/node1_master
	e.name = fmt.Sprint(e.stdName, "_", "master")
	return e
}

func (e *EtcdRegistry) AsSlave() *EtcdRegistry {
	e.name = fmt.Sprint(e.stdName, "_", "slave")
	return e
}

func (e *EtcdRegistry) LifecycleLease() <-chan clientv3.LeaseID {
	return e.leaseIdChan
}

func (e *EtcdRegistry) GetServiceMapping(name string) map[string]string {
	resp, err := e.cli.Get(context.Background(), EtcdPrefix.FmtRegistry(e.cfg.Group, name), clientv3.WithPrefix())
	if err != nil {
		registryLog.Errorf("get services: %s", err)
		return map[string]string{}
	}
	res := make(map[string]string, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		idx := bytes.LastIndexByte(kv.Key, '/')
		if idx < 0 {
			continue
		}
		sid, _, _ := bytes.Cut(kv.Key[idx+1:], []byte("_"))
		res[util.BytesToStr(sid)] = util.BytesToStr(kv.Value)
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

func (e *EtcdRegistry) MustRegister() *EtcdRegistry {
	if err := e.Register(); err != nil {
		panic(err)
	}
	return e
}

func (e *EtcdRegistry) Register() error {
	// skip if already register
	if e.leaseId != -1 {
		return nil
	}
	ctx := context.Background()
	var err error

	if e.leaseId, err = e.makeKvWithLease(ctx, e.Key(), e.addr); err != nil {
		return err
	}
	e.leaseIdChan <- e.leaseId
	var keepAlive <-chan *clientv3.LeaseKeepAliveResponse
	keepAlive, e.stopFn, err = e.keepaliveLease(e.leaseId)
	if err != nil {
		return err
	}
	//listen the heartbeat response
	go func() {
		defer graceful.Recover()
		for resp := range keepAlive {
			registryLog.Tracef("keepalive %s success (%d)", e.Key(), resp.TTL)
		}
		registryLog.Infof("stop keepalive %s", e.Key())
	}()
	registryLog.Infof("registry %s success", e.Key())
	return nil
}

func (e *EtcdRegistry) Unregister() error {
	registryLog.Tracef("manual unregister %s", e.Key())
	e.stopFn()
	if e.leaseId != -1 {
		ctx, cancel := context.WithTimeout(context.Background(), e.cfg.Timeout)
		defer cancel()
		_, err := e.cli.Delete(ctx, e.Key())
		if err != nil {
			return err
		}
		_, err = e.cli.Revoke(ctx, e.leaseId)
		if err != nil {
			return err
		}
		e.leaseId = -1
	}
	return nil
}

func (e *EtcdRegistry) makeKvWithLease(ctx context.Context, key string, value string) (clientv3.LeaseID, error) {
	//grant a lease
	lease, err := e.cli.Grant(ctx, int64(e.cfg.Interval.Seconds()))
	if err != nil {
		return -1, fmt.Errorf("Register interval heartbeat: grant lease error, %v", err)
	}
	//create a key with lease
	if _, err = e.cli.Put(ctx, key, value, clientv3.WithLease(lease.ID)); err != nil {
		return -1, fmt.Errorf("Register interval heartbeat: send heartbeat error, %v", err)
	}
	return lease.ID, nil
}

func (e *EtcdRegistry) keepaliveLease(id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, func(), error) {
	ctx, cancel := context.WithCancel(context.Background())
	ch, err := e.cli.KeepAlive(ctx, id)
	return ch, func() {
		defer cancel()
		e.stopFn = func() {}
	}, err
}
