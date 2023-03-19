package service

import (
	"common/cst"
	"common/graceful"
	"common/logs"
	"common/util"
	"context"
	"fmt"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var locatorLog = logs.New("object-locator")

type Locator struct {
	etcd *clientv3.Client
}

func NewLocator(etcd *clientv3.Client) *Locator {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := etcd.Get(ctx, cst.EtcdPrefix.LocationSubKey, clientv3.WithCountOnly())
	if err != nil {
		util.LogErr(err)
	}
	if err != nil || res.Count == 0 {
		_, err = etcd.Put(ctx, cst.EtcdPrefix.LocationSubKey, "")
		util.LogErr(err)
	}
	return &Locator{etcd}
}

// StartLocate start a thread to subscribe a special key on ETCD.
// reply object position to publisher
func (l *Locator) StartLocate(ip string) (cancel func()) {
	ctx := context.Background()
	ctx, cancel = context.WithCancel(ctx)
	ch := l.etcd.Watch(ctx, cst.EtcdPrefix.LocationSubKey)
	go func() {
		defer graceful.Recover()
		locatorLog.Info("Start listening locating message...")
		for {
			select {
			case <-ctx.Done():
				locatorLog.Info("Cancel listening locating message")
				return
			case resp, ok := <-ch:
				if !ok {
					locatorLog.Warn("Locate watching stop! Try watching again...")
					ch = l.etcd.Watch(ctx, cst.EtcdPrefix.LocationSubKey)
					break
				}
				if resp.Err() != nil {
					locatorLog.Error(resp.Err())
					break
				}
				for _, event := range resp.Events {
					go l.handlerLocate(event.Kv.Value, ip)
				}
			}
		}
	}()
	return cancel
}

// handlerLocate receive "hash.idx#key" response "ip#idx"
func (l *Locator) handlerLocate(message []byte, ip string) {
	defer graceful.Recover()
	tp := strings.Split(string(message), "#")
	if len(tp) != 2 {
		locatorLog.Errorf("Receive incorrect message %s", string(message))
		return
	}
	hash, respKey := tp[0], tp[1]
	locatorLog.Tracef("handler locating request: hash=%s, response to key %s", hash, respKey)
	if Exist(hash) {
		tp = strings.Split(hash, ".")
		if len(tp) != 2 {
			logs.Std().Errorf("Receive incorrect message %s", string(message))
			return
		}
		loc := fmt.Sprint(ip, "#", tp[1])
		_, err := l.etcd.Put(context.Background(), respKey, loc)
		if err != nil {
			locatorLog.Errorf("Put locate repsone on key %s error: %s", respKey, err)
			return
		}
		locatorLog.Debugf("put %s to key %s success", loc, respKey)
	}
}
