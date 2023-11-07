package timing_wheel

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (s *WaitGroupWrapper) AddDone(fun func()) {
	s.Add(1)

	go func() {
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

	go func() {
		timeWheel.spin()
	}()
	return timeWheel
}

func (w *TimingWheel) spin() {
	for {
		select {
		case bucket := <-w.queue.OfferC():
			w.currentTime.Store(bucket.expireTime)

			w.AddDone(func() {
				w.bucketClean(bucket)
			})

		case <-w.close:
			return
		}
	}
}

func (w *TimingWheel) AddAfter(duration time.Duration, fun func()) {
	task := &TimerTask{
		expireTime: TruncateTime(time.Now().UnixMilli()+duration.Milliseconds(), w.wheel.tick),
		fun:        fun,
	}

	w.addOrRun(task)
}

func (w *TimingWheel) addOrRun(task *TimerTask) {
	// 小于当前时间，直接运行
	currentTime := w.currentTime.Load()

	if task.expireTime < currentTime {
		w.AddDone(task.fun)
		return
	}

	wheel := w.wheel
	for {
		// 处在当前时间轮里
		if task.expireTime < currentTime+wheel.interval {
			// 找到 bucket
			index := (task.expireTime - currentTime) / wheel.tick

			w.Lock()
			bucket := wheel.buckets[index%wheel.wheelSize]
			// 新建 bucket
			if bucket == nil {
				bucket = NewBucket()
				wheel.buckets[index%wheel.wheelSize] = bucket
			}
			w.Unlock()

			// 放到 bucket 里
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
