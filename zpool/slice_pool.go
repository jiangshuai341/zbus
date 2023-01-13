package zpool

import (
	"math/bits"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	minBitSize uint32 = 6 // CPU cache line bitSize 64bit
	poolNum           = 20

	calibrateCallsThreshold = 42000
	maxPercentile           = 0.95
)

type slicePool[T any] struct {
	callCounter   [poolNum]uint64
	isCalibrating uint64

	defaultBitSize uint32
	maxBitSize     uint32

	pools [poolNum]sync.Pool
}

func (p *slicePool[T]) Get() (buf []T) {
	defaultBitSize := atomic.LoadUint32(&p.defaultBitSize)
	bufLen := 1 << defaultBitSize
	ptr, _ := p.pools[defaultBitSize].Get().(unsafe.Pointer)
	if ptr == nil {
		return make([]T, bufLen, bufLen)
	}
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Data = uintptr(ptr)
	sh.Len = bufLen
	sh.Cap = bufLen
	runtime.KeepAlive(ptr)
	return
}

func (p *slicePool[T]) Get2(size int) (buf []T) {
	idx := index(size)
	bitSize := uint32(idx) + minBitSize
	ptr, _ := p.pools[idx].Get().(unsafe.Pointer)
	if ptr == nil {
		return make([]T, 1<<bitSize)[:size]
	}
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Data = uintptr(ptr)
	sh.Len = size
	sh.Cap = 1 << bitSize
	runtime.KeepAlive(ptr)
	return
}

func (p *slicePool[T]) Put(buf []T) {
	size := cap(buf)
	if size == 0 || uint32(size) > 1<<poolNum+minBitSize || uint32(size) < 1<<minBitSize {
		return
	}
	idx := index(size)
	bitSize := uint32(idx) + minBitSize
	if size != 1<<bitSize { // this byte slice is not from Pool.Get()
		idx--
	}
	if atomic.AddUint64(&p.callCounter[idx], 1) > calibrateCallsThreshold {
		p.calibrate()
	}
	maxBitSize := int(atomic.LoadUint32(&p.maxBitSize))
	if maxBitSize == 0 || size <= maxBitSize {
		return
	}
	p.pools[idx].Put(unsafe.Pointer(&buf[:1][0]))
}

func (p *slicePool[T]) calibrate() {
	if !atomic.CompareAndSwapUint64(&p.isCalibrating, 0, 1) {
		return
	}

	a := make(callSizes, 0, poolNum)
	var callsSum uint64

	for i := uint32(0); i < poolNum; i++ {
		calls := atomic.SwapUint64(&p.callCounter[i], 0)
		callsSum += calls
		a = append(a, callSize{
			calls:   calls,
			bitSize: minBitSize + i,
		})
	}
	sort.Sort(a)

	defaultBitSize := a[0].bitSize
	maxBitSize := defaultBitSize

	maxSum := uint64(float64(callsSum) * maxPercentile)
	callsSum = 0
	for i := 0; i < poolNum; i++ {
		if callsSum > maxSum {
			break
		}
		callsSum += a[i].calls
		size := a[i].bitSize
		if size > maxBitSize {
			maxBitSize = size
		}
	}

	atomic.StoreUint32(&p.defaultBitSize, defaultBitSize)
	atomic.StoreUint32(&p.maxBitSize, maxBitSize)

	atomic.StoreUint64(&p.isCalibrating, 0)
}

type callSize struct {
	calls   uint64
	bitSize uint32
}

type callSizes []callSize

func (ci callSizes) Len() int {
	return len(ci)
}

func (ci callSizes) Less(i, j int) bool {
	return ci[i].calls > ci[j].calls
}

func (ci callSizes) Swap(i, j int) {
	ci[i], ci[j] = ci[j], ci[i]
}

func index(n int) int {
	n--
	n >>= minBitSize
	idx := bits.Len32(uint32(n))
	if idx >= poolNum {
		idx = poolNum - 1
	}
	return idx
}
