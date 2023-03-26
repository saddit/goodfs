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
	"sync"
	"time"
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

type Lifecycle struct {
	Client      clientv3.Lease
	mux         sync.RWMutex
	lease       clientv3.LeaseID
	ttl         time.Duration
	ctx         context.Context
	cancel      func()
	notifyLease []func(id clientv3.LeaseID)
}

func NewLifecycle(cli clientv3.Lease, ttl time.Duration) *Lifecycle {
	ctx, cancel := context.WithCancel(context.Background())
	return &Lifecycle{
		Client: cli,
		mux:    sync.RWMutex{},
		ttl:    ttl,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (lc *Lifecycle) notifying() {
	for _, f := range lc.notifyLease {
		go func(fn func(id clientv3.LeaseID)) {
			defer graceful.Recover()
			fn(lc.lease)
		}(f)
	}
}

func (lc *Lifecycle) Lease() clientv3.LeaseID {
	lc.mux.RLock()
	defer lc.mux.RUnlock()
	return lc.lease
}

func (lc *Lifecycle) DeadLoop() {
	defer graceful.Recover()
	for {
		// lock to update lease
		lc.mux.Lock()
		leaseResp, err := lc.Client.Grant(context.Background(), int64(lc.ttl.Seconds()))
		if err != nil {
			registryLog.Errorf("grant lifecycle lease err: %s", err)
			continue
		}
		lc.lease = leaseResp.ID
		lc.mux.Unlock()

		// notify lease change
		lc.notifying()

		// keepalive new lease
		response, err := lc.Client.KeepAlive(lc.ctx, lc.lease)
		if err != nil {
			registryLog.Errorf("could not keepalive for lifecycle lease %d, err: %s", lc.lease, err)
			continue
		}
		for r := range response {
			registryLog.Tracef("keepalive lease %d success: ttl=%d", r.ID, r.TTL)
		}

		// break loop if closed
		select {
		case <-lc.ctx.Done():
			_, _ = lc.Client.Revoke(context.Background(), lc.lease)
			registryLog.Infof("stop keepalive lifecycle lease")
			return
		default:
			registryLog.Warnf("lifecycle lease revoked, start grant a new lease")
		}
	}
}

func (lc *Lifecycle) Subscribe(fn func(id clientv3.LeaseID)) {
	lc.notifyLease = append(lc.notifyLease, fn)
}

func (lc *Lifecycle) Close() error {
	lc.cancel()
	return nil
}
