package conn_pool

import (
	"sort"
	"sync"
	"time"

	"github.com/yongpi/putil/plog"
)

type ConnState int

const (
	Idle ConnState = iota
	Activity
	Busy
)

type Closer interface {
	Close()
}

func NewConn[T Closer](conn T, balance *Balancer[T]) *Conn[T] {
	return &Conn[T]{connect: conn, balance: balance, state: Idle}
}

type Conn[T Closer] struct {
	sync.RWMutex
	count   int
	state   ConnState
	connect T
	balance *Balancer[T]
}

func (c *Conn[T]) use() {
	c.Lock()
	defer c.Unlock()

	c.count++
	if c.count < c.balance.markBusyCount {
		c.state = Activity
		return
	}

	c.state = Busy
	c.balance.incrBusyCount()
}

func (c *Conn[T]) Close() {
	c.Lock()
	defer c.Unlock()

	c.count--
	if c.count >= c.balance.markBusyCount {
		return
	}

	c.state = Activity
	c.balance.decrBusyCount()
	return
}

func (c *Conn[T]) GetConnect() T {
	return c.connect
}

func NewBalance[T Closer](options ...BalanceOption[T]) *Balancer[T] {
	cfg := &BalanceConfig[T]{
		core:          3,
		max:           5,
		markBusyCount: 5,
		lowLoadRatio:  0.65,
		checkDuration: 15 * time.Second,
	}

	for _, option := range options {
		option(cfg)
	}

	if cfg.newConn == nil {
		panic("new conn must need")
	}

	balance := &Balancer[T]{
		list:          make([]*Conn[T], 0),
		BalanceConfig: cfg,
	}

	go balance.release()
	return balance
}

func newBalance[T Closer](cfg *BalanceConfig[T]) *Balancer[T] {
	if cfg.newConn == nil {
		panic("new conn must need")
	}

	balance := &Balancer[T]{
		list:          make([]*Conn[T], 0),
		BalanceConfig: cfg,
	}

	go balance.release()
	return balance
}

type Balancer[T Closer] struct {
	sync.RWMutex
	*BalanceConfig[T]
	list  []*Conn[T]
	index int
}

func (b *Balancer[T]) IsHighLoad() bool {
	if float64(b.busyCount)/float64(len(b.list)) > b.lowLoadRatio {
		return true
	}

	return false
}

func (b *Balancer[T]) release() {
	ticker := time.Tick(b.checkDuration)

	for {
		select {
		case <-ticker:
			plog.Debugf("[Balancer] check release")

			// 连接池没到核心数，或者处于高负载，不释放连接
			if len(b.list) <= b.core || b.IsHighLoad() {
				continue
			}

			b.Lock()
			if len(b.list) == 0 {
				continue
			}

			// 清理之前先排序
			b.sort()
			// 数组最后一个连接连接数最少
			conn := b.list[len(b.list)-1]
			if conn.count > 0 {
				b.Unlock()
				continue
			}

			plog.Debugf("[Balancer] release conn = %#v", conn)

			conn.connect.Close()
			b.list = b.list[:len(b.list)-1]
			b.Unlock()
		}
	}
}

func (b *Balancer[T]) sort() {
	sort.Slice(b.list, func(i, j int) bool {
		return b.list[i].count > b.list[j].count
	})
}

func (b *Balancer[T]) incrBusyCount() {
	b.Lock()
	defer b.Unlock()

	b.busyCount++
	return
}

func (b *Balancer[T]) decrBusyCount() {
	b.Lock()
	defer b.Unlock()

	b.busyCount--
	return
}

func (b *Balancer[T]) newConnect(service string) *Conn[T] {
	connect := NewConn(b.newConn(service), b)

	plog.Debugf("[Balancer] new connect = %#v", connect)

	b.Lock()
	defer b.Unlock()
	b.list = append(b.list, connect)
	b.sort()
	return connect
}

func (b *Balancer[T]) Connect(service string) *Conn[T] {
	// 核心池没满，直接新建连接
	if len(b.list) < b.core {
		return b.newConnect(service)
	}

	// 核心池已满，并且没有到最大连接数
	// 判断是否高负载
	if len(b.list) < b.max && b.IsHighLoad() {
		return b.newConnect(service)
	}

	// 低负载或者达到最大连接数
	b.RLock()
	defer b.RUnlock()

	conn := b.list[b.index%len(b.list)]
	if conn.state != Busy {
		b.index++
		conn.use()

		plog.Debugf("[Balancer] use first not busy conn = %#v", conn)
		return conn
	}

	second := b.list[b.index+1%len(b.list)]
	if second.state != Busy {
		b.index++
		second.use()

		plog.Debugf("[Balancer] use second not busy conn = %#v", conn)
		return second
	}

	b.index++
	conn.use()

	plog.Debugf("[Balancer] use busy conn = %#v", conn)
	return conn
}
