package locate

import (
	"common/graceful"
	"common/logs"
	"context"
	"fmt"
	"objectserver/internal/usecase/service"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	LocationSubKey = "goodfs.location"
)

var logrus = logs.New("object-locator")

type Locator struct {
	etcd *clientv3.Client
}

func New(etcd *clientv3.Client) *Locator {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if res, _ := etcd.Get(ctx, LocationSubKey, clientv3.WithCountOnly()); res.Count == 0 {
		etcd.Put(ctx, LocationSubKey, "")
	}
	return &Locator{etcd}
}

//StartLocate 监听 etcd key 实现定位消息接收
func (l *Locator) StartLocate(ip string) {
	ctx := context.Background()
	ch := l.etcd.Watch(ctx, LocationSubKey)
	go func() {
		defer graceful.Recover()
		logrus.Info("Start listenning locating message...")
		for {
			resp, ok := <-ch
			if !ok {
				logrus.Warn("Locate watching stop! Try watching again...")
				ch = l.etcd.Watch(ctx, LocationSubKey)
				continue
			}
			if resp.Err() != nil {
				logrus.Error(resp.Err())
				continue
			}
			for _, event := range resp.Events {
				go l.handlerLocate(event.Kv.Value, ip)
			}
		}
	}()
}

// handlerLocate recieve "hash.idx#key" response "ip#idx"
func (l *Locator) handlerLocate(message []byte, ip string) {
	defer graceful.Recover()
	tp := strings.Split(string(message), "#")
	if len(tp) != 2 {
		logrus.Errorf("Receive incorrect message %s", string(message))
		return
	}
	hash, respKey := tp[0], tp[1]
	if service.Exist(hash) {
		tp = strings.Split(hash, ".")
		if len(tp) != 2 {
			logrus.Errorf("Receive incorrect message %s", string(message))
			return
		}
		_, err := l.etcd.Put(context.Background(), respKey, fmt.Sprint(ip, "#", tp[1]))
		if err != nil {
			logrus.Errorf("Put locate repsone on key %s error: %s", respKey, err)
			return
		}
	}
}
