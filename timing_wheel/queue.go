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

	go queue.watch()
	return queue
}

func (q *DelayBucketQueue) watch() {
	for {
		timeMs := time.Now().UnixMilli()

		q.Lock()
		bucket, delay := q.popTimeMs(timeMs)
		q.Unlock()

		if bucket != nil {
			q.c <- bucket
			continue
		}

		if bucket == nil && delay == 0 {
			select {
			case <-q.modify:
				continue
			}
		}

		if bucket != nil && delay != 0 {
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
	if len(q.priorityQueue.list) == 0 {
		return nil, 0
	}

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

	go func() {
		q.modify <- struct{}{}
	}()
}

func (q *DelayBucketQueue) OfferC() <-chan *Bucket {
	return q.c
}
