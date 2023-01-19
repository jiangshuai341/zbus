package coroutinepool

import (
	"sync"
	"sync/atomic"
	"testing"
)

var num int64 = 1
var lock sync.RWMutex

func BenchmarkRlock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lock.RLock()
		i++
		lock.RUnlock()
	}
}
func BenchmarkWlock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lock.Lock()
		i++
		lock.Unlock()
	}
}
func BenchmarkAtomic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		atomic.AddInt64(&num, 1)
	}
}
func BenchmarkNothing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		num++
	}
}
