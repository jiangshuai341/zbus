package lockfreequeue

import (
	"fmt"
	"runtime"
	"sync/atomic"
)

type Element struct {
	putCount uint32
	getCount uint32
	value    any
}

// LockFreeRingArray lock free Ring Queue Array
type LockFreeRingArray struct {
	capacity uint32
	capMod   uint32
	putPos   uint32
	getPos   uint32
	cache    []Element
}

func NewLockFreeRingArray(capacity uint32) *LockFreeRingArray {
	q := new(LockFreeRingArray)
	q.capacity = minQuantity(capacity)
	q.capMod = q.capacity - 1
	q.putPos = 0
	q.getPos = 0
	q.cache = make([]Element, q.capacity)
	for i := range q.cache {
		cache := &q.cache[i]
		cache.getCount = uint32(i)
		cache.putCount = uint32(i)
	}
	cache := &q.cache[0]
	cache.getCount = q.capacity
	cache.putCount = q.capacity
	return q
}

func (que *LockFreeRingArray) String() string {
	getPos := atomic.LoadUint32(&que.getPos)
	putPos := atomic.LoadUint32(&que.putPos)
	return fmt.Sprintf("Queue{capacity: %v, capMod: %v, putPos: %v, getPos: %v}",
		que.capacity, que.capMod, putPos, getPos)
}

func (que *LockFreeRingArray) Capaciity() uint32 {
	return que.capacity
}

func (que *LockFreeRingArray) Quantity() uint32 {
	var putPos, getPos uint32
	var quantity uint32
	getPos = atomic.LoadUint32(&que.getPos)
	putPos = atomic.LoadUint32(&que.putPos)

	if putPos >= getPos {
		quantity = putPos - getPos
	} else {
		quantity = que.capMod + (putPos - getPos)
	}

	return quantity
}

func (que *LockFreeRingArray) Put(val any) (ok bool, quantity uint32) {
	var putPos, putPosNew, getPos, posCnt uint32
	var cache *Element
	capMod := que.capMod

	getPos = atomic.LoadUint32(&que.getPos)
	putPos = atomic.LoadUint32(&que.putPos)

	if putPos >= getPos {
		posCnt = putPos - getPos
	} else {
		posCnt = capMod + (putPos - getPos)
	}

	if posCnt >= capMod-1 {
		runtime.Gosched()
		return false, posCnt
	}

	putPosNew = putPos + 1
	if !atomic.CompareAndSwapUint32(&que.putPos, putPos, putPosNew) {
		runtime.Gosched()
		return false, posCnt
	}

	cache = &que.cache[putPosNew&capMod]

	for {
		getCount := atomic.LoadUint32(&cache.getCount)
		putCount := atomic.LoadUint32(&cache.putCount)
		if putPosNew == putCount && getCount == putCount {
			cache.value = val
			atomic.AddUint32(&cache.putCount, que.capacity)
			return true, posCnt + 1
		} else {
			runtime.Gosched()
		}
	}
}

func (que *LockFreeRingArray) Puts(values []any) (puts, quantity uint32) {
	var putPos, putPosNew, getPos, posCnt, putCnt uint32
	capMod := que.capMod

	getPos = atomic.LoadUint32(&que.getPos)
	putPos = atomic.LoadUint32(&que.putPos)

	if putPos >= getPos {
		posCnt = putPos - getPos
	} else {
		posCnt = capMod + (putPos - getPos)
	}

	if posCnt >= capMod-1 {
		runtime.Gosched()
		return 0, posCnt
	}

	if capPuts, size := que.capacity-posCnt, uint32(len(values)); capPuts >= size {
		putCnt = size
	} else {
		putCnt = capPuts
	}
	putPosNew = putPos + putCnt

	if !atomic.CompareAndSwapUint32(&que.putPos, putPos, putPosNew) {
		runtime.Gosched()
		return 0, posCnt
	}

	for posNew, v := putPos+1, uint32(0); v < putCnt; posNew, v = posNew+1, v+1 {
		var cache *Element = &que.cache[posNew&capMod]
		for {
			getCount := atomic.LoadUint32(&cache.getCount)
			putCount := atomic.LoadUint32(&cache.putCount)
			if posNew == putCount && getCount == putCount {
				cache.value = values[v]
				atomic.AddUint32(&cache.putCount, que.capacity)
				break
			} else {
				runtime.Gosched()
			}
		}
	}
	return putCnt, posCnt + putCnt
}

func (que *LockFreeRingArray) Get() (val any, ok bool, quantity uint32) {
	var putPos, getPos, getPosNew, posCnt uint32
	var cache *Element
	capMod := que.capMod

	putPos = atomic.LoadUint32(&que.putPos)
	getPos = atomic.LoadUint32(&que.getPos)

	if putPos >= getPos {
		posCnt = putPos - getPos
	} else {
		posCnt = capMod + (putPos - getPos)
	}

	if posCnt < 1 {
		runtime.Gosched()
		return nil, false, posCnt
	}

	getPosNew = getPos + 1
	if !atomic.CompareAndSwapUint32(&que.getPos, getPos, getPosNew) {
		runtime.Gosched()
		return nil, false, posCnt
	}

	cache = &que.cache[getPosNew&capMod]

	for {
		getCount := atomic.LoadUint32(&cache.getCount)
		putCount := atomic.LoadUint32(&cache.putCount)
		if getPosNew == getCount && getCount == putCount-que.capacity {
			val = cache.value
			cache.value = nil
			atomic.AddUint32(&cache.getCount, que.capacity)
			return val, true, posCnt - 1
		} else {
			runtime.Gosched()
		}
	}
}

func (que *LockFreeRingArray) Gets(values []any) (gets, quantity uint32) {
	var putPos, getPos, getPosNew, posCnt, getCnt uint32
	capMod := que.capMod

	putPos = atomic.LoadUint32(&que.putPos)
	getPos = atomic.LoadUint32(&que.getPos)

	if putPos >= getPos {
		posCnt = putPos - getPos
	} else {
		posCnt = capMod + (putPos - getPos)
	}

	if posCnt < 1 {
		runtime.Gosched()
		return 0, posCnt
	}

	if size := uint32(len(values)); posCnt >= size {
		getCnt = size
	} else {
		getCnt = posCnt
	}
	getPosNew = getPos + getCnt

	if !atomic.CompareAndSwapUint32(&que.getPos, getPos, getPosNew) {
		runtime.Gosched()
		return 0, posCnt
	}

	for posNew, v := getPos+1, uint32(0); v < getCnt; posNew, v = posNew+1, v+1 {
		var cache *Element = &que.cache[posNew&capMod]
		for {
			getCount := atomic.LoadUint32(&cache.getCount)
			putCount := atomic.LoadUint32(&cache.putCount)
			if posNew == getCount && getCount == putCount-que.capacity {
				values[v] = cache.value
				cache.value = nil
				getCount = atomic.AddUint32(&cache.getCount, que.capacity)
				break
			} else {
				runtime.Gosched()
			}
		}
	}

	return getCnt, posCnt - getCnt
}

// round 到最近的2的倍数
func minQuantity(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

func Delay(z int) {
	for x := z; x > 0; x-- {
	}
}
