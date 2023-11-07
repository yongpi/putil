package timing_wheel

import "container/list"

type TimerTask struct {
	bucket     *Bucket
	expireTime int64
	fun        func()
	element    *list.Element
}
