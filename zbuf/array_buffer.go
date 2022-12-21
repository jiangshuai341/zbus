package zbuf

import (
	"github.com/jiangshuai341/zbus/zpool"
	"reflect"
	"unsafe"
)

type ArrayBuffers struct {
	buf        [][]byte
	size       int
	blockSize  int
	prefixSize int
}

func NewArraryBuffers(numPrefix int, size int, blockSize int) *ArrayBuffers {
	ret := &ArrayBuffers{}
	ret.prefixSize = numPrefix
	ret.blockSize = blockSize
	ret.reserve(size)
	return ret
}
func (a *ArrayBuffers) BufferWithPrefix() *[][]byte {
	return &a.buf
}
func (a *ArrayBuffers) BufferWithoutPrefix() [][]byte {
	return a.buf[a.prefixSize:]
}

func (a *ArrayBuffers) reserve(newsize int) {
	mallocSize := newsize - a.size
	if mallocSize <= 0 {
		return
	}
	if a.buf == nil {
		var temp []byte
		if a.blockSize > 0 {
			temp = zpool.Get2(a.blockSize)
		} else {
			temp = zpool.Get()
		}
		a.blockSize = len(temp)
		a.buf = make([][]byte, a.prefixSize, newsize/a.blockSize+a.prefixSize)
		a.buf = append(a.buf, temp)
		a.size += a.blockSize
	}
	for ; mallocSize > 0; mallocSize -= a.blockSize {
		a.buf = append(a.buf, zpool.Get2(a.blockSize))
		a.size += a.blockSize
	}
}
func (a *ArrayBuffers) MoveTemp(num int) *[][]byte {
	if num <= 0 || num > a.size {
		return nil
	}
	var ret = make([][]byte, num/a.blockSize+1)
	for i := 0; num > 0; i++ {
		ret[i] = a.buf[a.prefixSize+i]
		a.buf[a.prefixSize+i] = zpool.Get2(a.blockSize)
		if num < a.blockSize {
			sh := (*reflect.SliceHeader)(unsafe.Pointer(&ret[i]))
			sh.Len = num
		}
		num -= a.blockSize
	}
	return &ret
}

func (a *ArrayBuffers) GetPrefix() [][]byte {
	return a.buf[:a.prefixSize]
}
