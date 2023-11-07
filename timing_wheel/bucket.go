package timing_wheel

import (
	"container/list"
	"sync"
	"sync/atomic"
)

type Bucket struct {
	sync.Mutex
	expireTime int64
	list       *list.List
	a          atomic.Int64
}

func NewBucket() *Bucket {
	return &Bucket{
		expireTime: -1,
		list:       list.New(),
	}
}

func (b *Bucket) Clear() []*TimerTask {
	data := make([]*TimerTask, 0)

	b.Lock()
	defer b.Unlock()

	node := b.list.Front()
	for node != nil {
		next := node.Next()
		b.list.Remove(node)

		task := node.Value.(*TimerTask)
		task.element = nil
		task.bucket = nil
		data = append(data, task)

		node = next
	}

	// 重置过期时间，则 bucket 可以再次被加入到延迟队列中
	b.expireTime = -1
	return data
}

func (b *Bucket) Push(task *TimerTask) bool {
	b.Lock()
	defer b.Unlock()

	element := b.list.PushBack(task)
	task.element = element
	task.bucket = b

	if b.expireTime == -1 {
		b.expireTime = task.expireTime
		return true
	}

	return false
}

type BucketHeap struct {
	list []*Bucket
}

func (h *BucketHeap) Len() int {
	return len(h.list)
}

func (h *BucketHeap) Less(i, j int) bool {
	return h.list[i].expireTime < h.list[j].expireTime
}

func (h *BucketHeap) Swap(i, j int) {
	h.list[i], h.list[j] = h.list[j], h.list[i]
}

func (h *BucketHeap) Push(x any) {
	h.list = append(h.list, x.(*Bucket))
}

func (h *BucketHeap) Pop() any {
	item := h.list[0]
	h.list = h.list[1:]

	return item
}
