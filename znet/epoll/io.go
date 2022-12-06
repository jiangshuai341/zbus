package epoll

import (
	"syscall"
	"unsafe"
)

var _zero uintptr

type Iovec struct {
	Base *byte
	Len  uint64
}

func Readv(fd int, buffers [][]byte) (int, error) {
	iovecs := bytes2iovec(buffers)
	if len(iovecs) == 0 {
		return 0, nil
	}
	_p0 := unsafe.Pointer(&iovecs[0])
	n, _, err := syscall.Syscall(syscall.SYS_READV, uintptr(fd), uintptr(_p0), uintptr(len(iovecs)))
	return int(n), err
}
func Writev(fd int, buffers [][]byte) (int, error) {
	var ptr unsafe.Pointer
	iovecs := bytes2iovec(buffers)
	if len(iovecs) == 0 {
		return 0, nil
	}
	ptr = unsafe.Pointer(&iovecs[0])
	n, _, err := syscall.Syscall(syscall.SYS_WRITEV, uintptr(fd), uintptr(ptr), uintptr(len(iovecs)))
	return int(n), err
}

func bytes2iovec(bs [][]byte) []Iovec {
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
