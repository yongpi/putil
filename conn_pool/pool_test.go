package conn_pool

import (
	"fmt"
	"sync/atomic"
	"testing"
)

type TestCloser struct {
	ID int64
}

func (t *TestCloser) Close() {
	fmt.Printf("[TestCloser] close id = %d\n", t.ID)
}

var idGen atomic.Int64

func NewTestCloser() *TestCloser {
	return &TestCloser{ID: idGen.Add(1)}
}

func TestNewBalancePool(t *testing.T) {
	pool := NewBalancePool[*TestCloser](SetNewConn(func(serviceName string) *TestCloser {
		return NewTestCloser()
	}))

	c1 := pool.Conn("test")
	c2 := pool.Conn("test")
	c3 := pool.Conn("test")
	c4 := pool.Conn("test")
	c5 := pool.Conn("test")
	c6 := pool.Conn("test")
	c7 := pool.Conn("test2")

	c1.Close()
	c2.Close()
	c3.Close()
	c4.Close()
	c5.Close()
	c6.Close()
	c7.Close()

}
