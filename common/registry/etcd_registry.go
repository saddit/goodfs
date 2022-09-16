package registry

import (
	. "common/constrant"
	"common/graceful"
	"common/logs"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var log = logs.New("etcd-registry")

type EtcdRegistry struct {
	cli       *clientv3.Client
	cfg       Config
	leaseId   clientv3.LeaseID
	group     string
	stdName   string // be like metaserver_150013
	name      string // be like metaserver_150013_master
	localAddr string
	stopFn    func()
}

func NewEtcdRegistry(kv *clientv3.Client, cfg Config, localAddr string) *EtcdRegistry {
	ts := time.Now().Unix()
	return &EtcdRegistry{
		cli:       kv,
		cfg:       cfg,
		leaseId:   -1,
		group:     cfg.Group,
		stdName:   fmt.Sprint(cfg.Name, "_", ts),
		name:      fmt.Sprint(cfg.Name, "_", ts),
		localAddr: localAddr,
		stopFn:    nil,
	}
}

func (e *EtcdRegistry) Key() string {
	return EtcdPrefix.FmtRegistry(e.group, e.name)
}

func (e *EtcdRegistry) AsMaster() *EtcdRegistry {
	// goodfs/metaserver_123123_master
	e.name = fmt.Sprint(e.stdName, "_", "master")
	return e
}

func (e *EtcdRegistry) AsSlave() *EtcdRegistry {
	e.name = fmt.Sprint(e.stdName, "_", "slave")
	return e
}

func (e *EtcdRegistry) MustRegister() *EtcdRegistry {
	if err := e.Register(); err != nil {
		panic(err)
	}
	return e
}

func (e *EtcdRegistry) Register() error {
	ctx := context.Background()
	var err error
	//init registered key
	if e.leaseId, err = e.makeKvWithLease(ctx, e.Key(), e.localAddr); err != nil {
		return err
	}
	var kach <-chan *clientv3.LeaseKeepAliveResponse
	kach, e.stopFn, err = e.keepaliveLease(ctx, e.leaseId)
	if err != nil {
		return err
	}
	//listen the heartbeat response
	go func() {
		defer graceful.Recover()
		for resp := range kach {
			log.Tracef("keepalive %s success (%d)", e.Key(), resp.TTL)
		}
		log.Infof("stop keepalive %s", e.Key())
	}()
	log.Infof("registry %s success", e.Key())
	return nil
}

func (e *EtcdRegistry) Unregister() error {
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
	}
	return nil
}

func (e *EtcdRegistry) makeKvWithLease(ctx context.Context, key, value string) (clientv3.LeaseID, error) {
	//grant a lease
	ctx2, cancel2 := context.WithTimeout(ctx, e.cfg.Timeout)
	defer cancel2()
	lease, err := e.cli.Grant(ctx2, int64(e.cfg.Interval.Seconds()))
	if err != nil {
		return -1, fmt.Errorf("Register interval heartbeat: grant lease error, %v", err)
	}
	//create a key with lease
	ctx3, cancel3 := context.WithTimeout(ctx, e.cfg.Timeout)
	defer cancel3()
	if _, err := e.cli.Put(ctx3, key, value, clientv3.WithLease(lease.ID)); err != nil {
		return -1, fmt.Errorf("Register interval heartbeat: send heartbeat error, %v", err)
	}
	return lease.ID, nil
}

func (e *EtcdRegistry) keepaliveLease(ctx context.Context, id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, func(), error) {
	ctx2, cancel := context.WithCancel(ctx)
	ch, err := e.cli.KeepAlive(ctx2, id)
	return ch, cancel, err
}
