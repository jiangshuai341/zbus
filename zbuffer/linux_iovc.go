//go:build linux

package zbuffer

import (
	"github.com/jiangshuai341/zbus/znet/tcp-linux/epoll"
	"github.com/jiangshuai341/zbus/zpool/slicepool"
	"math/bits"
	"reflect"
	"sync"
	"unsafe"
)

var iovcPool sync.Pool

func GetIovcSlice() []epoll.Iovec {
	return iovcPool.Get().([]epoll.Iovec)
}
func PutIovcSlice(sl []epoll.Iovec) {
	iovcPool.Put(sl)
}
func (llb *LinkListBuffer) PeekToIovecs(iocvArrPtr *[]epoll.Iovec) {
	iter := llb.head
	*iocvArrPtr = (*iocvArrPtr)[:cap(*iocvArrPtr)]
	var i = 0
	for ; i < cap(*iocvArrPtr); i++ {
		if iter == nil {
			break
		}
		(*iocvArrPtr)[i] = epoll.Slice2Iovec(iter.buf)
		iter = iter.next
	}
	*iocvArrPtr = (*iocvArrPtr)[:i]
}

type IovcArray struct {
	buf        []epoll.Iovec
	size       int
	blockSize  int
	prefixSize int
}

func NewIocvArr(numPrefix int, size int, blockSize int) *IovcArray {
	ret := &IovcArray{}
	ret.prefixSize = numPrefix
	if blockSize < 64 {
		blockSize = 64
	}
	blockSize--
	ret.blockSize = 1 << bits.Len32(uint32(blockSize))
	ret.reserve(size)
	return ret
}

func (ia *IovcArray) BufferWithPrefix() []epoll.Iovec {
	return ia.buf
}
func (ia *IovcArray) BufferWithoutPrefix() []epoll.Iovec {
	return ia.buf[ia.prefixSize:]
}

func (ia *IovcArray) reserve(newsize int) {
	if newsize-ia.size <= 0 {
		return
	}
	if ia.buf == nil {
		ia.buf = make([]epoll.Iovec, ia.prefixSize, newsize/ia.blockSize+ia.prefixSize)
	}
	for ia.size < newsize {
		ia.buf = append(ia.buf, epoll.Slice2Iovec(slicepool.GetBuffer2(ia.blockSize)))
		ia.size += ia.blockSize
	}
}
func (ia *IovcArray) MoveTemp(num int) *[][]byte {
	if num <= 0 || num > ia.size {
		return nil
	}
	var ret = make([][]byte, num/ia.blockSize+1)

	for i := 0; num > 0; i++ {
		index := ia.prefixSize + i
		ia.buf[index].MoveToSlice((*reflect.SliceHeader)(unsafe.Pointer(&ret[i])))

		ia.buf[index].Len = uint64(ia.blockSize)
		ia.buf[index].Base = &(slicepool.GetBuffer2(ia.blockSize)[0])

		if num < ia.blockSize {
			sh := (*reflect.SliceHeader)(unsafe.Pointer(&ret[i]))
			sh.Len = num
		}

		num -= ia.blockSize
	}
	return &ret
}

var _zero uintptr

func (ia *IovcArray) SetPrefix(preFix ...[]byte) {
	if len(preFix) > ia.prefixSize {
		return
	}
	for i, v := range preFix {
		ia.buf[i].Len = uint64(len(v))
		if v != nil && len(v) > 0 {
			ia.buf[i].Base = &v[0]
		} else {
			ia.buf[i].Base = (*byte)(unsafe.Pointer(&_zero))
		}
	}
}
