package registry

import (
	"common/graceful"
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
	"time"
)

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
			time.Sleep(5 * time.Second)
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
			time.Sleep(5 * time.Second)
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
