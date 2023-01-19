package lockfreequeue

import (
	"sync/atomic"
	"unsafe"
)

// LockFreeList is a simple, fast, and practical non-blocking and concurrent queue with no lock.
type LockFreeList struct {
	head   unsafe.Pointer
	tail   unsafe.Pointer
	length int32
}

type node struct {
	value any
	next  unsafe.Pointer
}

func NewLockFreeList() *LockFreeList {
	n := unsafe.Pointer(&node{})
	return &LockFreeList{head: n, tail: n}
}

func (q *LockFreeList) Enqueue(task any) {
	n := &node{value: task}
retry:
	tail := load(&q.tail)
	next := load(&tail.next)
	if tail == load(&q.tail) {
		if next == nil {
			if cas(&tail.next, next, n) {
				cas(&q.tail, tail, n)
				atomic.AddInt32(&q.length, 1)
				return
			}
		} else {
			cas(&q.tail, tail, next)
		}
	}
	goto retry
}

func (q *LockFreeList) Dequeue() any {
retry:
	head := load(&q.head)
	tail := load(&q.tail)
	next := load(&head.next)
	if head == load(&q.head) {
		if head == tail {
			if next == nil {
				return nil
			}
			cas(&q.tail, tail, next)
		} else {
			task := next.value
			if cas(&q.head, head, next) {
				atomic.AddInt32(&q.length, -1)
				return task
			}
		}
	}
	goto retry
}

func (q *LockFreeList) IsEmpty() bool {
	return atomic.LoadInt32(&q.length) == 0
}

func load(p *unsafe.Pointer) (n *node) {
	return (*node)(atomic.LoadPointer(p))
}

func cas(p *unsafe.Pointer, old, new *node) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
