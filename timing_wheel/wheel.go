package timing_wheel

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yongpi/putil/plog"
)

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (s *WaitGroupWrapper) SafeRun(fun func()) {
	s.Add(1)

	go func() {
		defer func() {
			// 如果发生 panic 则执行不到 Done 方法，需要 recover
			if err := recover(); err != nil {
				plog.Errorf("[WaitGroupWrapper] fun panic!, err = %#v", err)
				s.Done()
			}
		}()

		fun()
		s.Done()
	}()
}

type Wheel struct {
	tick      int64
	wheelSize int64
	interval  int64
	buckets   []*Bucket
	overflow  *Wheel
}

type TimingWheel struct {
	WaitGroupWrapper
	sync.Mutex
	wheel       *Wheel
	currentTime atomic.Int64
	queue       *DelayBucketQueue
	close       chan any
}

func NewTimingWheel(tick time.Duration, wheelSize int64) *TimingWheel {
	if tick < time.Millisecond {
		panic(fmt.Sprintf("timing wheel tick too small, must >= 1ms"))
	}

	tickMs := int64(tick / time.Millisecond)

	firstWheel := &Wheel{
		tick:      tickMs,
		wheelSize: wheelSize,
		interval:  tickMs * wheelSize,
		buckets:   make([]*Bucket, wheelSize),
	}

	queue := NewDelayBucketQueue()
	cc := make(chan any, 1)

	timeWheel := &TimingWheel{
		wheel: firstWheel,
		queue: queue,
		close: cc,
	}

	timeWheel.currentTime.Store(TruncateTime(time.Now().UnixMilli(), tickMs))

	// 时间轮开始转动
	go func() {
		timeWheel.spin()
	}()
	return timeWheel
}

func (w *TimingWheel) spin() {
	for {
		select {
		// 延迟队列中是否存在数据
		case bucket := <-w.queue.OfferC():
			// 更新时间轮当前时间戳
			w.currentTime.Store(bucket.expireTime)

			// 清理桶中的任务：执行或者到新的时间轮的新桶中
			w.SafeRun(func() {
				w.bucketClean(bucket)
			})

		case <-w.close:
			return
		}
	}
}

func (w *TimingWheel) AddAfter(duration time.Duration, fun func()) {
	task := &TimerTask{
		// 根据时间轮的 tick 参数，计算出来对应的放到时间轮中的时间
		expireTime: TruncateTime(time.Now().UnixMilli()+duration.Milliseconds(), w.wheel.tick),
		fun:        fun,
	}

	w.addOrRun(task)
}

func (w *TimingWheel) addOrRun(task *TimerTask) {
	currentTime := w.currentTime.Load()

	// 小于当前时间，直接运行
	if task.expireTime < currentTime {
		w.SafeRun(task.fun)
		return
	}

	// 大于当前时间，遍历时间轮选择加入
	wheel := w.wheel
	for {
		// 处在当前时间轮里
		if task.expireTime < currentTime+wheel.interval {
			// 找到 bucket
			index := (task.expireTime - currentTime) / wheel.tick

			// 加锁查找时间轮中的桶，有并发的情况，需要加锁查找和新建桶
			w.Lock()
			bucket := wheel.buckets[index%wheel.wheelSize]
			// 新建 bucket
			if bucket == nil {
				bucket = NewBucket()
				wheel.buckets[index%wheel.wheelSize] = bucket
			}
			w.Unlock()

			// 放到 bucket 里，如果桶原来的过期时间为 -1，说明桶没有加入到延迟队列
			// 则加入到延迟队列中
			if bucket.Push(task) {
				w.queue.Push(bucket)
			}
			return
		}

		// 不在当前时间轮里，往后找
		w.Lock()
		if wheel.overflow == nil {
			wheel.overflow = &Wheel{
				tick:      wheel.interval,
				wheelSize: wheel.wheelSize,
				interval:  wheel.interval * wheel.wheelSize,
				buckets:   make([]*Bucket, wheel.wheelSize),
			}
		}
		w.Unlock()

		// 注意时间也要累加
		currentTime += wheel.interval
		wheel = wheel.overflow
	}
}

func (w *TimingWheel) bucketClean(bucket *Bucket) {
	list := bucket.Clear()
	for _, task := range list {
		// 注意这里，清理出来的 task 假如没到执行时间，则重新加入新的 bucket 中
		w.addOrRun(task)
	}
}

func (w *TimingWheel) Close() {
	w.close <- struct{}{}
	w.WaitGroupWrapper.Wait()
}

func TruncateTime(currentTime, tickMs int64) int64 {
	if tickMs == 0 {
		return currentTime
	}

	return currentTime - currentTime%tickMs
}
