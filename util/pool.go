package util

// @Time : 2020年3月13日12:27:33
// @Author : Lemyhello
// @Desc: 通用连接池

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	//ErrMaxActiveConnReached 连接池超限
	ErrMaxActiveConnReached = errors.New("MaxActiveConnReached")
)

// Config 连接池相关配置
type Config[T interface{}] struct {
	//连接池中拥有的最小连接数
	InitialCap int
	//最大并发存活连接数
	MaxCap int
	//最大空闲连接
	MaxIdle int
	//生成连接的方法
	Factory func() (*T, error)
	//关闭连接的方法
	Close func(*T) error
	//检查连接是否有效的方法
	Ping func(*T) error
	//连接最大空闲时间，超过该事件则将失效
	IdleTimeout time.Duration
}

// channelPool 存放连接信息
type channelPool[T interface{}] struct {
	mu                       sync.RWMutex
	conns                    chan *idleConn[T]
	factory                  func() (*T, error)
	close                    func(*T) error
	ping                     func(*T) error
	idleTimeout, waitTimeOut time.Duration
	maxActive                int
	openingConns             int
}

type idleConn[T interface{}] struct {
	conn *T
	t    time.Time
}

var (
	//ErrClosed 连接池已经关闭Error
	ErrClosed = errors.New("pool is closed")
)

// Pool 基本方法
type Pool[T interface{}] interface {
	Get() (*T, error)

	Put(*T) error

	Close(*T) error

	Release()

	Len() int
}

// NewChannelPool 初始化连接
func NewChannelPool[T interface{}](poolConfig *Config[T]) (Pool[T], error) {
	if !(poolConfig.InitialCap <= poolConfig.MaxIdle && poolConfig.MaxCap >= poolConfig.MaxIdle && poolConfig.InitialCap >= 0) {
		return nil, errors.New("invalid capacity settings")
	}
	if poolConfig.Factory == nil {
		return nil, errors.New("invalid factory func settings")
	}
	if poolConfig.Close == nil {
		return nil, errors.New("invalid close func settings")
	}

	c := &channelPool[T]{
		conns:        make(chan *idleConn[T], poolConfig.MaxIdle),
		factory:      poolConfig.Factory,
		close:        poolConfig.Close,
		idleTimeout:  poolConfig.IdleTimeout,
		maxActive:    poolConfig.MaxCap,
		openingConns: poolConfig.InitialCap,
	}

	if poolConfig.Ping != nil {
		c.ping = poolConfig.Ping
	}

	for i := 0; i < poolConfig.InitialCap; i++ {
		conn, err := c.factory()
		if err != nil {
			c.Release()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.conns <- &idleConn[T]{conn: conn, t: time.Now()}
	}

	return c, nil
}

// getConns 获取所有连接
func (c *channelPool[T]) getConns() chan *idleConn[T] {
	c.mu.Lock()
	conns := c.conns
	c.mu.Unlock()
	return conns
}

// Get 从pool中取一个连接
func (c *channelPool[T]) Get() (*T, error) {
	conns := c.getConns()
	if conns == nil {
		return nil, ErrClosed
	}
	for {
		select {
		case wrapConn := <-conns:
			if wrapConn == nil {
				return nil, ErrClosed
			}
			//判断是否超时，超时则丢弃
			if timeout := c.idleTimeout; timeout > 0 {
				if wrapConn.t.Add(timeout).Before(time.Now()) {
					//丢弃并关闭该连接
					c.Close(wrapConn.conn)
					continue
				}
			}
			//判断是否失效，失效则丢弃，如果用户没有设定 ping 方法，就不检查
			if c.ping != nil {
				if err := c.Ping(wrapConn.conn); err != nil {
					c.Close(wrapConn.conn)
					continue
				}
			}
			return wrapConn.conn, nil
		default:
			c.mu.Lock()
			defer c.mu.Unlock()
			if c.openingConns >= c.maxActive {
				return nil, ErrMaxActiveConnReached
			}
			if c.factory == nil {
				return nil, ErrClosed
			}
			conn, err := c.factory()
			if err != nil {
				return nil, err
			}
			c.openingConns++
			return conn, nil
		}
	}
}

// Put 将连接放回pool中
func (c *channelPool[T]) Put(conn *T) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	c.mu.Lock()

	if c.conns == nil {
		c.mu.Unlock()
		return c.Close(conn)
	}

	select {
	case c.conns <- &idleConn[T]{conn: conn, t: time.Now()}:
		c.mu.Unlock()
		return nil
	default:
		c.mu.Unlock()
		//连接池已满，直接关闭该连接
		return c.Close(conn)
	}
}

// Close 关闭单条连接
func (c *channelPool[T]) Close(conn *T) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.close == nil {
		return nil
	}
	c.openingConns--
	return c.close(conn)
}

// Ping 检查单条连接是否有效
func (c *channelPool[T]) Ping(conn *T) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}
	return c.ping(conn)
}

// Release 释放连接池中所有连接
func (c *channelPool[T]) Release() {
	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.factory = nil
	c.ping = nil
	closeFun := c.close
	c.close = nil
	c.mu.Unlock()

	if conns == nil {
		return
	}

	close(conns)
	for wrapConn := range conns {
		//log.Printf("Type %v\n",reflect.TypeOf(wrapConn.conn))
		closeFun(wrapConn.conn)
	}
}

// Len 连接池中已有的连接
func (c *channelPool[T]) Len() int {
	return len(c.getConns())
}
