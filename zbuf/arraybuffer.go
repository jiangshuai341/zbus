package zbuf

import (
	"github.com/jiangshuai341/zbus/zpool"
	"reflect"
	"unsafe"
)

type ArrayBuffer struct {
	buf       [][]byte
	size      int
	blockSize int
}

func (a *ArrayBuffer) Buffer() *[][]byte {
	return &a.buf
}
func (a *ArrayBuffer) Reserve(newsize int) {
	mallocSize := newsize - a.size
	if mallocSize <= 0 {
		return
	}
	if a.buf == nil {
		temp := zpool.Get()
		a.blockSize = len(temp)
		a.buf = make([][]byte, newsize/a.blockSize)
		a.buf[0] = temp
		a.size += a.blockSize
	}
	for ; mallocSize > 0; mallocSize -= a.blockSize {
		a.buf = append(a.buf, zpool.Get2(a.blockSize))
		a.size += a.blockSize
	}
}
func (a *ArrayBuffer) MoveTemp(num int) *[][]byte {
	if num > a.size {
		panic("moveTempHead arg error please check")
	}
	var ret = make([][]byte, num/a.blockSize, num/a.blockSize+1)
	for i := 0; num > 0; i++ {
		ret[i] = a.buf[i]
		a.buf[i] = zpool.Get2(a.blockSize)
		num -= a.blockSize
		if num < 0 {
			sh := (*reflect.SliceHeader)(unsafe.Pointer(&ret[i]))
			sh.Len = num + a.blockSize
		}
	}
	return &ret
}
