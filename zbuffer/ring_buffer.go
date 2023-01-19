package zbuffer

import (
	"errors"
	"github.com/jiangshuai341/zbus/toolkit"
	"github.com/jiangshuai341/zbus/zpool/slicepool"
)

const (
	DefaultBufferSize = 1024 // 1KB
)

var ErrIsEmpty = errors.New("ring-buffer is empty")
var ErrIsFull = errors.New("ring-buffer is full")

// RingBuffer is a circular buffer And implement io.ReaderWriter interface.
type RingBuffer struct {
	buf     []byte
	size    int
	r       int // next position to read
	w       int // next position to write
	isEmpty bool
}

func NewRingBuffer(size int) *RingBuffer {
	if size == 0 {
		return &RingBuffer{isEmpty: true}
	}
	return &RingBuffer{
		buf:     slicepool.GetBuffer2(size),
		size:    size,
		isEmpty: true,
	}
}

func (rb *RingBuffer) PeekDataSpace() (head []byte, tail []byte) {
	if rb.IsEmpty() {
		return
	}

	if rb.w > rb.r {
		head = rb.buf[rb.r:rb.w]
		return
	}

	head = rb.buf[rb.r:]
	if rb.w != 0 {
		tail = rb.buf[:rb.w]
	}

	return
}

func (rb *RingBuffer) PeekFreeSpace() (head []byte, tail []byte) {
	if rb.IsEmpty() {
		return rb.buf, nil
	}
	if rb.IsFull() {
		return nil, nil
	}

	if rb.w < rb.r {
		head = rb.buf[rb.w:rb.r]
		return
	}

	head = rb.buf[rb.w:]
	if rb.w != 0 {
		tail = rb.buf[:rb.r]
	}

	return
}

// Discard Grow FreeSpace
func (rb *RingBuffer) Discard(n int) int {
	if n <= 0 {
		return 0
	}

	LengthData := rb.LengthData()
	if n < LengthData {
		rb.r = (rb.r + n) % rb.size
		return n
	}
	rb.Reset()
	return LengthData
}

// DataSpaceGrow 标记写入了更多数据
func (rb *RingBuffer) DataSpaceGrow(n int) int {
	if n <= 0 {
		return 0
	}
	LengthFree := rb.LengthFree()
	if n > LengthFree {
		n = LengthFree
	}
	rb.w = (rb.w + n) % rb.size
	rb.isEmpty = false
	return n
}

func (rb *RingBuffer) WriteToSlice(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	if rb.LengthData() == 0 {
		return 0, ErrIsEmpty
	}
	head, tail := rb.PeekDataSpace()
	n1 := copy(p, head)

	if n1 < len(head) {
		return n1, nil
	}

	n2 := copy(p[n1:], tail)

	return n1 + n2, nil
}

// ReadFromSlice DataSpace Grow
func (rb *RingBuffer) ReadFromSlice(p []byte) int {
	if len(p) == 0 {
		return 0
	}
	if len(p) > rb.LengthFree() {
		rb.grow(rb.size + len(p) - rb.LengthFree())
	}

	head, tail := rb.PeekFreeSpace()

	n1 := copy(head, p)
	if n1 >= len(p) {
		return n1
	}
	n2 := copy(tail, p[n1:])
	rb.isEmpty = false
	return n1 + n2
}

// ReadFromString writes the contents of the string s to buffer, which accepts a slice of bytes.
func (rb *RingBuffer) ReadFromString(s string) int {
	return rb.ReadFromSlice(toolkit.StringToBytes(s))
}

// LengthData returns the length of available bytes to read.
func (rb *RingBuffer) LengthData() int {
	if rb.isEmpty {
		return 0
	}
	temp := rb.w - rb.r
	if temp > 0 {
		return temp
	}
	return temp + rb.size
}

// LengthFree returns the length of available bytes to write.
func (rb *RingBuffer) LengthFree() int {
	return rb.size - rb.LengthData()
}

func (rb *RingBuffer) Cap() int {
	return rb.size
}

// Bytes DataSpace
func (rb *RingBuffer) Bytes() []byte {

	if rb.IsEmpty() {
		return nil
	}
	head, tail := rb.PeekDataSpace()

	return append(head, tail...)
}

// IsFull tells if this ring-buffer is full.
func (rb *RingBuffer) IsFull() bool {
	return rb.r == rb.w && !rb.isEmpty
}

// IsEmpty tells if this ring-buffer is empty.
func (rb *RingBuffer) IsEmpty() bool {
	return rb.isEmpty
}

// Reset the read pointer and write pointer to zero.
func (rb *RingBuffer) Reset() {
	rb.isEmpty = true
	rb.r, rb.w = 0, 0
}

func (rb *RingBuffer) grow(newCap int) {
	if newCap <= DefaultBufferSize {
		newCap = DefaultBufferSize
	}
	newBuf := slicepool.GetBuffer2(newCap)
	oldLen := rb.LengthData()
	_, _ = rb.WriteToSlice(newBuf)
	slicepool.PutBuffer(rb.buf)
	rb.buf = newBuf
	rb.r = 0
	rb.w = oldLen
	rb.size = newCap
	if rb.w > 0 {
		rb.isEmpty = false
	}
}
