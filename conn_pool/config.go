package conn_pool

import "time"

type BalanceConfig[T Closer] struct {
	core          int
	max           int
	busyCount     int
	markBusyCount int
	lowLoadRatio  float64
	newConn       func(serviceName string) T
	checkDuration time.Duration
}

type BalanceOption[T Closer] func(cfg *BalanceConfig[T])

func Core[T Closer](core int) BalanceOption[T] {
	return func(cfg *BalanceConfig[T]) {
		cfg.core = core
	}
}

func Max[T Closer](max int) BalanceOption[T] {
	return func(cfg *BalanceConfig[T]) {
		cfg.max = max
	}
}

func BusyCount[T Closer](busyCount int) BalanceOption[T] {
	return func(cfg *BalanceConfig[T]) {
		cfg.busyCount = busyCount
	}
}

func MarkBusyCount[T Closer](markBusyCount int) BalanceOption[T] {
	return func(cfg *BalanceConfig[T]) {
		cfg.markBusyCount = markBusyCount
	}
}

func LowLoadRatio[T Closer](lowLoadRatio float64) BalanceOption[T] {
	return func(cfg *BalanceConfig[T]) {
		cfg.lowLoadRatio = lowLoadRatio
	}
}

func SetNewConn[T Closer](newConn func(serviceName string) T) BalanceOption[T] {
	return func(cfg *BalanceConfig[T]) {
		cfg.newConn = newConn
	}
}

func CheckDuration[T Closer](checkDuration time.Duration) BalanceOption[T] {
	return func(cfg *BalanceConfig[T]) {
		cfg.checkDuration = checkDuration
	}
}
