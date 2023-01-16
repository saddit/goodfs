package performance

import (
	"common/collection/set"
	"common/graceful"
	"common/util"
	"common/util/slices"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const etcdPrefix = "performance-data"
const maxSize = 1000 * 100

// EtcdStore implements Store using ETCD to storage data.
// it will not save all the data written.
// maximum write capacity limit to 10,000 entries and only the specified actions are allowed to be written to.
type EtcdStore struct {
	kv             clientv3.KV
	allowedActions set.Set
	mux            sync.Locker
}

func (es *EtcdStore) getKeyPrefix(action, kind string) string {
	return fmt.Sprintf("%s/%s/%s", etcdPrefix, action, kind)
}

func (es *EtcdStore) removeOldKeys(limit int) error {
	resp, err := es.kv.Get(context.Background(), etcdPrefix,
		clientv3.WithPrefix(),
		clientv3.WithKeysOnly(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		clientv3.WithLimit(int64(limit)),
	)
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return nil
	}
	_, err = es.kv.Delete(context.Background(),
		util.BytesToStr(slices.First(resp.Kvs).Key),
		clientv3.WithRange(util.BytesToStr(slices.Last(resp.Kvs).Key)),
	)
	return err
}

func (es *EtcdStore) checkSize(addedSize int) error {
	es.mux.Lock()
	defer es.mux.Unlock()
	curSize, err := es.Size("", "")
	if err != nil {
		return err
	}
	if exceed := int(curSize) + addedSize - maxSize; exceed > 0 {
		if err = es.removeOldKeys(exceed); err != nil {
			return fmt.Errorf("remove elder err: %w", err)
		}
	}
	return nil
}

func (es *EtcdStore) Put(pm []*Perform) error {
	var allows []*Perform
	for _, data := range pm {
		if es.allowedActions.Contains(data.Action) {
			allows = append(allows, data)
		}
	}
	if err := es.checkSize(len(allows)); err != nil {
		return err
	}
	tx := es.kv.Txn(context.Background())
	for _, data := range allows {
		key := fmt.Sprintf("%s/%d", es.getKeyPrefix(data.Action, data.KindOf), time.Now().UnixMilli())
		tx.Then(clientv3.OpPut(key, data.Cost.String()))
	}
	_, err := tx.Commit()
	return err
}

func (es *EtcdStore) Get(kindOf, action string) (res []*Perform, err error) {
	allowedActions := set.To[string](es.allowedActions)
	if action != "" {
		allowedActions = []string{action}
	}
	resCh := make(chan []*Perform, 1)
	go func() {
		defer graceful.Recover()
		defer close(resCh)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		dg := util.NewDoneGroup()
		defer dg.Close()
		for _, act := range allowedActions {
			dg.Todo()
			go func(a string) {
				defer dg.Done()
				resp, err := es.kv.Get(ctx, es.getKeyPrefix(a, kindOf), clientv3.WithPrefix())
				if err != nil {
					dg.Error(err)
					return
				}
				arr := make([]*Perform, 0, len(resp.Kvs))
				for _, kv := range resp.Kvs {
					key := util.BytesToStr(kv.Key)
					idx := strings.LastIndexByte(key, '/')
					if idx < 0 {
						continue
					}
					arr = append(arr, &Perform{
						KindOf: key[idx+1:],
						Action: action,
						Cost:   time.Duration(util.ToInt64(util.BytesToStr(kv.Value))),
					})
				}
				resCh <- arr
			}(act)
		}
		err = dg.WaitUntilError()
	}()
	for arr := range resCh {
		res = append(res, arr...)
	}
	return
}

func (es *EtcdStore) Clear(kindOf, action string) (err error) {
	allowedActions := set.To[string](es.allowedActions)
	if action != "" {
		allowedActions = []string{action}
	}
	tx := es.kv.Txn(context.Background())
	for _, act := range allowedActions {
		tx.Then(clientv3.OpDelete(es.getKeyPrefix(act, kindOf), clientv3.WithPrefix()))
	}
	_, err = tx.Commit()
	return
}

func (es *EtcdStore) Size(kindOf, action string) (total int64, err error) {
	allowedActions := set.To[string](es.allowedActions)
	if action != "" {
		allowedActions = []string{action}
	}
	go func() {
		defer graceful.Recover()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		dg := util.NewDoneGroup()
		defer dg.Close()
		for _, act := range allowedActions {
			dg.Todo()
			go func(a string) {
				defer dg.Done()
				resp, err := es.kv.Get(ctx, es.getKeyPrefix(a, kindOf), clientv3.WithPrefix(), clientv3.WithCountOnly())
				if err != nil {
					dg.Error(err)
					return
				}
				atomic.AddInt64(&total, resp.Count)
			}(act)
		}
		err = dg.WaitUntilError()
	}()
	return
}

func (es *EtcdStore) Average(string, string) ([]*Perform, error) {
	panic("not implement Average")
}

func (es *EtcdStore) Sum(string, string) ([]*Perform, error) {
	panic("not implement Sum")
}

func NewEtcdStore(client *clientv3.Client, allowedActions []string) Store {
	sess, err := concurrency.NewSession(client, concurrency.WithTTL(15))
	if err != nil {
		util.PanicErr(fmt.Errorf("init etcd locker fail: %w", err))
	}
	return AvgSumStore(&EtcdStore{
		kv:             client,
		allowedActions: set.OfString(allowedActions),
		mux:            concurrency.NewLocker(sess, etcdPrefix),
	})
}
