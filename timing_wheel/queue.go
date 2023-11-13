package timing_wheel

import (
	"sync"
	"time"
)

type DelayBucketQueue struct {
	sync.Mutex
	priorityQueue *BucketHeap
	c             chan *Bucket
	modify        chan any
}

func NewDelayBucketQueue() *DelayBucketQueue {
	priority := &BucketHeap{list: make([]*Bucket, 0)}
	bc := make(chan *Bucket)
	wc := make(chan any)

	queue := &DelayBucketQueue{
		priorityQueue: priority,
		c:             bc,
		modify:        wc,
	}

	// 检查延迟队列的变化
	go queue.watch()
	return queue
}

func (q *DelayBucketQueue) watch() {
	for {
		// 根据当前时间，找到过期的 bucket
		timeMs := time.Now().UnixMilli()
		bucket, delay := q.popTimeMs(timeMs)

		// 存在，说明该桶需要被执行
		if bucket != nil {
			q.c <- bucket
			continue
		}

		// 延迟时间为 0，说明没有桶需要等待执行，阻塞
		if bucket == nil && delay == 0 {
			select {
			case <-q.modify:
				continue
			}
		}

		// 延迟时间不为 0，说明存在桶，并且还没到过期时间
		// 需要阻塞一段时间再次查询
		if bucket == nil && delay != 0 {
			select {
			case <-time.After(time.Millisecond * time.Duration(timeMs)):
				continue
			case <-q.modify:
				continue
			}
		}
	}
}

func (q *DelayBucketQueue) popTimeMs(timeMs int64) (*Bucket, int64) {
	q.Lock()
	defer q.Unlock()

	if len(q.priorityQueue.list) == 0 {
		return nil, 0
	}

	// 优先级队列，头部为时间最小桶
	head := q.priorityQueue.list[0]
	if head.expireTime >= timeMs {
		return head, 0
	}

	return nil, head.expireTime - timeMs
}

func (q *DelayBucketQueue) Push(bucket *Bucket) {
	q.Lock()
	q.priorityQueue.Push(bucket)
	q.Unlock()

	// 延迟队列发生了变化，异步发送信号
	go func() {
		q.modify <- struct{}{}
	}()
}

func (q *DelayBucketQueue) OfferC() <-chan *Bucket {
	return q.c
}
