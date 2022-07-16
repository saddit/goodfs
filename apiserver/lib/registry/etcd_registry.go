package registry

import (
	"apiserver/config"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdRegistry struct {
	*clientv3.Client
	cfg       config.RegistryConfig
	leaseId   clientv3.LeaseID
	key       string
	localAddr string
}

func NewEtcdRegistry(kv *clientv3.Client, cfg config.RegistryConfig, localAddr string) *EtcdRegistry {
	return &EtcdRegistry{
		kv, cfg, -1,
		fmt.Sprintf("%s/%s_%d", cfg.Group, cfg.Name, time.Now().UnixMilli()),
		localAddr,
	}
}

func (e *EtcdRegistry) Register() error {
	ctx := context.Background()
	var err error

	//init registered key
	if e.leaseId, err = e.makeKvWithLease(ctx, e.localAddr); err != nil {
		return err
	}

	//keepalive the lease
	kach, err := e.keepaliveLease(ctx, e.leaseId)
	if err != nil {
		return err
	}

	//listen the hearbeat response
	go func() {
		for resp := range kach {
			logrus.Infof("keepalive %s success (%d)", e.localAddr, resp.TTL)
		}
		logrus.Infof("stop keepalive %s", e.localAddr)
	}()

	return nil
}

func (e *EtcdRegistry) Unregister() error {
	if e.leaseId != -1 {
		ctx, cancel := context.WithTimeout(context.Background(), e.cfg.Timeout)
		defer cancel()

		_, err := e.Delete(ctx, e.key)
		if err != nil {
			return err
		}
		_, err = e.Revoke(ctx, e.leaseId)
		if err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("Unregister failed")
	}
}

func (e *EtcdRegistry) makeKvWithLease(ctx context.Context, addr string) (clientv3.LeaseID, error) {
	//grant a lease
	ctx2, cancel2 := context.WithTimeout(ctx, e.cfg.Timeout)
	defer cancel2()
	lease, err := e.Grant(ctx2, e.cfg.Interval.Milliseconds())
	if err != nil {
		return -1, fmt.Errorf("Register interval heartbeat: grant lease error, %v", err)
	}

	//create a key with lease
	ctx3, cancel3 := context.WithTimeout(ctx, e.cfg.Timeout)
	defer cancel3()
	if _, err := e.Put(ctx3, e.key, addr, clientv3.WithLease(lease.ID)); err != nil {
		return -1, fmt.Errorf("Register interval heartbeat: send heartbeat error, %v", err)
	}
	return lease.ID, nil
}

func (e *EtcdRegistry) keepaliveLease(ctx context.Context, id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	ctx2, cancel := context.WithTimeout(ctx, e.cfg.Timeout)
	defer cancel()
	return e.KeepAlive(ctx2, id)
}
