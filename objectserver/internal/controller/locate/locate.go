package locate

import (
	"common/graceful"
	"context"
	"fmt"
	"objectserver/internal/usecase/service"
	"strings"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Locator struct {
	etcd *clientv3.Client
}

func New(etcd *clientv3.Client) *Locator {
	return &Locator{etcd}
}

//StartLocate 监听 etcd key 实现定位消息接收
func (l *Locator) StartLocate(ip string) {
	key := fmt.Sprint("locate.", ip)
	ctx := context.Background()
	if _, err := l.etcd.Put(ctx, key, ""); err != nil {
		panic(err)
	}
	ch := l.etcd.Watch(ctx, key)
	go func() {
		defer graceful.Recover()
		for {
			select {
			case resp, ok := <-ch:
				if !ok {
					logrus.Warn("Locate watching stop! Try watching again...")
					ch = l.etcd.Watch(ctx, key)
					break
				}
				if resp.Err() != nil {
					logrus.Error(resp.Err())
					break
				}
				for _, event := range resp.Events {
					go l.handlerLocate(event.Kv.Value, ip)
				}
			}
		}
	}()
}

func (l *Locator) handlerLocate(message []byte, ip string) {
	defer graceful.Recover()
	tp := strings.Split(string(message), "#")
	if len(tp) != 2 {
		logrus.Errorf("Receive incorrect message %s", string(message))
		return
	}
	hash, respKey := tp[0], tp[1]
	if service.Exist(hash) {
		_, err := l.etcd.Put(context.Background(), respKey, ip)
		if err != nil {
			logrus.Errorf("Put locate repsone on key %s error: %s", respKey, err)
			return
		}
	}
}
