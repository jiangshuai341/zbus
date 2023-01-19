package epoll

import (
	"reflect"
	"syscall"
	"unsafe"
)

var _zero uintptr

type Iovec struct {
	Base *byte
	Len  uint64
}

func Readv(fd int, iovecs []Iovec) (int, error) {
	if len(iovecs) == 0 {
		return 0, nil
	}
	_p0 := unsafe.Pointer(&iovecs[0])
	n, _, err := syscall.Syscall(syscall.SYS_READV, uintptr(fd), uintptr(_p0), uintptr(len(iovecs)))
	if err == 0 {
		return int(n), nil
	}
	return int(n), err
}
func Writev(fd int, iovecs []Iovec) (int, error) {
	if len(iovecs) == 0 {
		return 0, nil
	}
	ptr := unsafe.Pointer(&iovecs[0])
	n, _, err := syscall.Syscall(syscall.SYS_WRITEV, uintptr(fd), uintptr(ptr), uintptr(len(iovecs)))
	if err == 0 {
		return int(n), nil
	}
	return int(n), err
}

func Slices2Iovec(bs [][]byte) []Iovec {
	iovecs := make([]Iovec, len(bs))
	for i, b := range bs {
		iovecs[i].Len = uint64(len(b))
		if len(b) > 0 {
			iovecs[i].Base = &b[0]
		} else {
			iovecs[i].Base = (*byte)(unsafe.Pointer(&_zero))
		}
	}
	return iovecs
}

func Slice2Iovec(bs []byte) Iovec {
	ret := Iovec{Len: uint64(len(bs))}
	if len(bs) == 0 {
		ret.Base = (*byte)(unsafe.Pointer(&_zero))
	} else {
		ret.Base = &bs[0]
	}
	return ret
}

func (i *Iovec) MoveToSlice(header *reflect.SliceHeader) {
	header.Len = int(i.Len)
	header.Data = uintptr(unsafe.Pointer(i.Base))
	header.Cap = int(i.Len)
	i.Len = 0
	i.Base = nil
}
