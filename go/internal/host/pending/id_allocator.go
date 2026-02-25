package pending

import "sync/atomic"

type IDAllocator struct {
	next atomic.Int64
}

func NewIDAllocator(start int64) *IDAllocator {
	a := &IDAllocator{}
	a.next.Store(start)
	return a
}

func (a *IDAllocator) Next() int64 {
	return a.next.Add(1)
}
