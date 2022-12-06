package lockfreequeue

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueuePutGet(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	const (
		isPrintf = false
	)

	cnt := 10000
	sum := 0
	start := time.Now()
	var putD, getD time.Duration
	for i := 0; i <= runtime.NumCPU()*4; i++ {
		sum += i * cnt
		put, get := testQueuePutGet(t, i, cnt)
		putD += put
		getD += get
	}
	end := time.Now()
	use := end.Sub(start)
	op := use / time.Duration(sum)
	t.Logf("Grp: %d, Times: %d, use: %v, %v/op", runtime.NumCPU()*4, sum, use, op)
	t.Logf("Put: %d, use: %v, %v/op", sum, putD, putD/time.Duration(sum))
	t.Logf("Get: %d, use: %v, %v/op", sum, getD, getD/time.Duration(sum))
}

func testQueuePutGet(t *testing.T, grp, cnt int) (
	put time.Duration, get time.Duration) {
	var wg sync.WaitGroup
	var id int32
	wg.Add(grp)
	q := NewLockFreeRingArray(1024 * 1024)
	start := time.Now()
	for i := 0; i < grp; i++ {
		go func(g int) {
			defer wg.Done()
			for j := 0; j < cnt; j++ {
				val := fmt.Sprintf("Node.%d.%d.%d", g, j, atomic.AddInt32(&id, 1))
				ok, _ := q.Put(&val)
				for !ok {
					time.Sleep(time.Microsecond)
					ok, _ = q.Put(&val)
				}
			}
		}(i)
	}
	wg.Wait()
	end := time.Now()
	put = end.Sub(start)

	wg.Add(grp)
	start = time.Now()
	for i := 0; i < grp; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < cnt; {
				_, ok, _ := q.Get()
				if !ok {
					runtime.Gosched()
				} else {
					j++
				}
			}
		}()
	}
	wg.Wait()
	end = time.Now()
	get = end.Sub(start)
	if q := q.Quantity(); q != 0 {
		t.Errorf("Grp:%v, Quantity Error: [%v] <>[%v]", grp, q, 0)
	}
	return put, get
}
