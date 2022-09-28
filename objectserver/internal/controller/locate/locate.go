package locate

import (
	"common/graceful"
	"common/logs"
	"common/util"
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

var log = logs.New("object-locator")

type Locator struct {
	etcd *clientv3.Client
}

func New(etcd *clientv3.Client) *Locator {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if res, _ := etcd.Get(ctx, LocationSubKey, clientv3.WithCountOnly()); res.Count == 0 {
		_, err := etcd.Put(ctx, LocationSubKey, "")
		util.LogErr(err)
	}
	return &Locator{etcd}
}

//StartLocate 监听 etcd key 实现定位消息接收, 执行返回的方法以终止
func (l *Locator) StartLocate(ip string) (cancel func()) {
	ctx := context.Background()
	ctx, cancel = context.WithCancel(ctx)
	ch := l.etcd.Watch(ctx, LocationSubKey)
	go func() {
		defer graceful.Recover()
		log.Info("Start listening locating message...")
		for {
			select {
			case <-ctx.Done():
				log.Info("Cancel listening locating message")
				return
			case resp, ok := <-ch:
				if !ok {
					log.Warn("Locate watching stop! Try watching again...")
					ch = l.etcd.Watch(ctx, LocationSubKey)
					break
				}
				if resp.Err() != nil {
					log.Error(resp.Err())
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
		logs.Std().Errorf("Receive incorrect message %s", string(message))
		return
	}
	hash, respKey := tp[0], tp[1]
	logs.Std().Debugf("handler locating request: hash=%s, response to key %s", hash, respKey)
	if service.Exist(hash) {
		tp = strings.Split(hash, ".")
		if len(tp) != 2 {
			logs.Std().Errorf("Receive incorrect message %s", string(message))
			return
		}
		loc := fmt.Sprint(ip, "#", tp[1])
		_, err := l.etcd.Put(context.Background(), respKey, loc)
		if err != nil {
			logs.Std().Errorf("Put locate repsone on key %s error: %s", respKey, err)
			return
		}
		logs.Std().Debugf("put %s to key %s success", loc, respKey)
	}
}
