package conn_pool

import (
	"sort"
	"sync"
	"time"
)

type ConnState int

const (
	Idle ConnState = iota
	Activity
	Busy
	Close
)

type Closer interface {
	Close()
}

func NewConn[T Closer](conn T, busyCount int) *Conn[T] {
	return &Conn[T]{connect: conn, markBusyCount: busyCount}
}

type Conn[T Closer] struct {
	sync.RWMutex
	count         int
	state         ConnState
	connect       T
	markBusyCount int
}

func (c *Conn[T]) use() {
	c.Lock()
	defer c.Unlock()

	c.count++
	if c.count >= c.markBusyCount {
		c.state = Busy
	}
}

func (c *Conn[T]) Close() {
	c.Lock()
	defer c.Unlock()

	c.count--
	if c.count < c.markBusyCount {
		c.state = Activity
	}
}

func (c *Conn[T]) GetConnect() T {
	return c.connect
}

func NewBalance[T Closer](serviceName string, core, max, markBusyCount int, lowLoadRatio float64, newFun func(serviceName string) T, checkDuration time.Duration) *Balancer[T] {
	balance := &Balancer[T]{
		list:          make([]*Conn[T], 0),
		core:          core,
		max:           max,
		markBusyCount: markBusyCount,
		lowLoadRatio:  lowLoadRatio,
		serviceName:   serviceName,
		newConn:       newFun,
		checkDuration: checkDuration,
	}

	go balance.Release()
	return balance
}

type Balancer[T Closer] struct {
	sync.RWMutex
	list          []*Conn[T]
	core          int
	max           int
	busyCount     int
	markBusyCount int
	lowLoadRatio  float64
	serviceName   string
	newConn       func(serviceName string) T
	checkDuration time.Duration
	index         int
}

func (b *Balancer[T]) IsHighLoad() bool {
	if float64(b.busyCount)/float64(len(b.list)) > b.lowLoadRatio {
		return true
	}

	return false
}

func (b *Balancer[T]) Release() {
	ticker := time.Tick(b.checkDuration)

	for {
		select {
		case <-ticker:
			if len(b.list) <= b.core || b.IsHighLoad() {
				continue
			}

			b.Lock()
			if len(b.list) == 0 {
				continue
			}
			conn := b.list[len(b.list)-1]
			if conn.count > 0 {
				b.Unlock()
				continue
			}

			conn.connect.Close()
			b.list = b.list[:len(b.list)-1]
			b.Unlock()
		}
	}
}

func (b *Balancer[T]) Sort() {
	sort.Slice(b.list, func(i, j int) bool {
		return b.list[i].count > b.list[j].count
	})
}

func (b *Balancer[T]) newConnect() *Conn[T] {
	connect := NewConn(b.newConn(b.serviceName), b.markBusyCount)
	b.Lock()
	defer b.Unlock()
	b.list = append(b.list, connect)
	b.Sort()
	return connect
}

func (b *Balancer[T]) Connect() *Conn[T] {
	// 核心池没满，直接新建连接
	if len(b.list) < b.core {
		return b.newConnect()
	}

	// 核心池已满，并且没有到最大连接数
	// 判断是否高负载
	if len(b.list) < b.max && b.IsHighLoad() {
		return b.newConnect()
	}

	// 低负载或者达到最大连接数
	b.RLock()
	defer b.RUnlock()

	conn := b.list[b.index%len(b.list)]
	if conn.state != Busy {
		b.index++
		return conn
	}

	second := b.list[b.index+1%len(b.list)]
	if second.state != Busy {
		b.index++
		return second
	}

	b.index++
	return conn
}
