package conn_pool

import (
	"sync"
	"time"
)

type BalancePool[T Closer] struct {
	sync.Mutex
	pool map[string]*Balancer[T]
	cfg  *BalanceConfig[T]
}

func NewBalancePool[T Closer](options ...BalanceOption[T]) *BalancePool[T] {
	pool := &BalancePool[T]{pool: make(map[string]*Balancer[T])}

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

	pool.cfg = cfg

	return pool
}

func (p *BalancePool[T]) Conn(service string) *Conn[T] {
	p.Lock()
	defer p.Unlock()

	balance, ok := p.pool[service]
	if !ok {
		balance = newBalance(p.cfg)
		p.pool[service] = balance
	}

	return balance.Connect(service)
}
