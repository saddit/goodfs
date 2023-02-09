package performance

import (
	"common/graceful"
	"common/logs"
	"common/util/slices"
	"context"
	"fmt"
	"sync"
	"time"
)

var (
	localStore  Store
	remoteStore Store
)

var pmLog = logs.New("performance-collector")

func SetLocalStore(s Store) {
	localStore = s
}

func SetRemoteStore(s Store) {
	remoteStore = s
}

func getStore(st StoreType) Store {
	switch st {
	case None:
		return NoneStore()
	case Local:
		if localStore == nil {
			panic("local store not set")
		}
		return localStore
	case Remote:
		if remoteStore == nil {
			panic("remote store not set")
		}
		return remoteStore
	default:
		panic("invalid store type")
	}
}

type Collector interface {
	Put(action string, kindOf string, cost time.Duration) error
	PutAsync(action string, kindOf string, cost time.Duration)
	Store() Store
	Flush() error
	Close() error
}

type pmCollector struct {
	store        Store
	conf         *Config
	memData      []*Perform
	lastSaveTime time.Time
	mux          sync.Locker
	stopAutoSave func()
}

func NewCollector(cfg *Config) Collector {
	if !cfg.Enable {
		return &noneCollector{NoneStore()}
	}
	c := &pmCollector{
		conf:         cfg,
		store:        getStore(cfg.Store),
		memData:      make([]*Perform, 0, cfg.MaxInMemory),
		lastSaveTime: time.Time{},
		mux:          &sync.Mutex{},
	}
	c.startAutoFlush()
	return c
}

func (c *pmCollector) Store() Store {
	return c.store
}

func (c *pmCollector) PutAsync(action string, kindOf string, cost time.Duration) {
	logs.Std().DebugFn(func() []any {
		stack := graceful.GetLimitStacks(6, 1)
		return []any{fmt.Sprintf("performance: [%s-%s] spend %s:\n%s", kindOf, action, cost, stack)}
	})
	go func() {
		defer graceful.Recover()
		if err := c.Put(action, kindOf, cost); err != nil {
			logs.Std().Errorf("put performance err: %s", err)
		}
	}()
}

func (c *pmCollector) Put(action string, kindOf string, cost time.Duration) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.memData = append(c.memData, &Perform{
		KindOf: kindOf,
		Action: action,
		Cost:   cost,
	})
	if len(c.memData) > c.conf.MaxInMemory {
		if c.conf.FlushWhenReached {
			if err := c.Flush(); err != nil {
				return err
			}
		} else {
			slices.RemoveFirst(&c.memData)
		}
	}
	return nil
}

// Flush save in-memory data to Store and clear memory buffer. it's goroutine unsafe.
func (c *pmCollector) Flush() error {
	c.lastSaveTime = time.Now()
	if err := c.store.Put(c.memData); err != nil {
		return err
	}
	slices.Clear(&c.memData)
	return nil
}

func (c *pmCollector) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.Flush(); err != nil {
		return err
	}
	if c.stopAutoSave != nil {
		c.stopAutoSave()
	}
	return nil
}

func (c *pmCollector) startAutoFlush() {
	go func() {
		defer graceful.Recover()
		tk := time.NewTicker(c.conf.FlushInterval)
		ctx, cancel := context.WithCancel(context.Background())
		c.stopAutoSave = cancel
		flushFn := func() {
			c.mux.Lock()
			defer c.mux.Unlock()
			if time.Since(c.lastSaveTime) < c.conf.FlushInterval {
				return
			}
			if err := c.Flush(); err != nil {
				pmLog.Errorf("auto flush fail: %s", err)
			}
		}
		for {
			select {
			case <-tk.C:
				flushFn()
			case <-ctx.Done():
				pmLog.Info("stop auto flush")
				return
			}
		}
	}()
}

type noneCollector struct {
	s Store
}

func (noneCollector) PutAsync(action string, kindOf string, cost time.Duration) {
}

func (noneCollector) Put(string, string, time.Duration) error {
	return nil
}

func (n noneCollector) Store() Store {
	return n.s
}

func (noneCollector) Flush() error {
	return nil
}

func (noneCollector) Close() error {
	return nil
}
