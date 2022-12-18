package zbuf

import (
	"errors"
	"github.com/jiangshuai341/zbus/zpool"
	"io"
	"reflect"
	"unsafe"
)

type node struct {
	buf  []byte
	next *node
}

func (b *node) len() int {
	return len(b.buf)
}

// LinkListBuffer is a linked list of node.
type LinkListBuffer struct {
	head  *node
	tail  *node
	size  int
	bytes int
}

func NewLinkListBuffer() *LinkListBuffer {
	return &LinkListBuffer{
		head:  nil,
		tail:  nil,
		size:  0,
		bytes: 0,
	}
}

// NewBytesFromPool 必不为空
func (llb *LinkListBuffer) NewBytesFromPool(len int) []byte {
	return zpool.Get2(len)
}

// ListLength 链表长度
func (llb *LinkListBuffer) ListLength() int {
	return llb.size
}

// ByteLength 所有可读的字节数
func (llb *LinkListBuffer) ByteLength() int {
	return llb.bytes
}

// IsEmpty 是否为空
func (llb *LinkListBuffer) IsEmpty() bool {
	return llb.head == nil
}

// Reset 删除所有元素
func (llb *LinkListBuffer) Reset() {
	for b := llb.pop(); b != nil; b = llb.pop() {
		zpool.Put(b.buf)
	}
	llb.head = nil
	llb.tail = nil
	llb.size = 0
	llb.bytes = 0
}

func (llb *LinkListBuffer) Emplace(num int, ret *[][]byte) {
	*ret = (*ret)[:0]
	if num <= 0 {
		return
	}
	for num > 0 {
		temp := zpool.Get()
		*ret = append(*ret, temp)
		llb.PushNoCopy(&temp)
		num -= len(temp)
	}
}
func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
func discardHead(buf *[]byte, num int) {
	if buf == nil || len(*buf) <= num {
		return
	}
	copy(*buf, (*buf)[num:])
	sh := (*reflect.SliceHeader)(unsafe.Pointer(buf))
	sh.Len = sh.Len - num
}
func (llb *LinkListBuffer) Peek(needNum int, ret *[][]byte) {
	if needNum <= 0 {
		needNum = llb.bytes
	}
	for iter := llb.head; iter != nil; iter = iter.next {
		*ret = append(*ret, iter.buf[:min(iter.len(), needNum)])
		if needNum -= iter.len(); needNum <= 0 {
			break
		}
	}
}

//func (llb *LinkListBuffer) PeekAndDiscard(needNum int) []byte {
//	if needNum <= 0 {
//		needNum = llb.bytes
//	}
//	var ret []byte
//	for iter := llb.head; iter != nil; iter = iter.next {
//		ret = append(ret, iter.buf[0:min(needNum, len(iter.buf))]...)
//		llb.DiscardBytes(min(needNum, len(iter.buf)))
//		if needNum -= iter.len(); needNum <= 0 {
//			break
//		}
//	}
//	return ret
//}

func (llb *LinkListBuffer) PeekInt64() int64 {
	if llb.bytes < 8 {
		return 0
	}
	return *(*int64)(unsafe.Pointer(&(llb.head.buf[0])))
}

func (llb *LinkListBuffer) PeekInt32() int32 {
	if llb.bytes < 4 {
		return 0
	}
	return *(*int32)(unsafe.Pointer(&(llb.head.buf[0])))
}

func (llb *LinkListBuffer) PeekInt16() int16 {
	if llb.bytes < 2 {
		return 0
	}
	return *(*int16)(unsafe.Pointer(&(llb.head.buf[0])))
}
func (llb *LinkListBuffer) PeekInt8() int8 {
	if llb.bytes < 1 {
		return 0
	}
	return *(*int8)(unsafe.Pointer(&(llb.head.buf[0])))
}

// DiscardBytes removes some nodes based on n bytes.
func (llb *LinkListBuffer) DiscardBytes(n int) (discarded int) {
	if n <= 0 {
		return
	}
	for n != 0 {
		b := llb.pop()
		if b == nil {
			break
		}
		if n < b.len() {
			discardHead(&b.buf, n)
			discarded += n
			llb.pushFront(b)
			break
		}
		n -= b.len()
		discarded += b.len()
		zpool.Put(b.buf)
	}
	return
}

// DiscardNodeNum removes some nodes based on n NodeNum.
func (llb *LinkListBuffer) DiscardNodeNum(n int) (discarded int) {
	if n <= 0 {
		return
	}
	for ; n != 0 && llb.size != 0; n-- {
		node := llb.pop()
		if node == nil {
			break
		}
		discarded++
		zpool.Put(node.buf)
	}
	return
}

func (llb *LinkListBuffer) Push(p []byte) {
	n := len(p)
	if n == 0 {
		return
	}
	b := zpool.Get2(n)
	copy(b, p)
	llb.pushBack(&node{buf: b})
}

func (llb *LinkListBuffer) PushNoCopy(p *[]byte) {
	if p == nil || len(*p) == 0 {
		return
	}
	llb.pushBack(&node{buf: *p})
}
func (llb *LinkListBuffer) PushsNoCopy(p *[][]byte) {
	if p == nil {
		return
	}
	for i := range *p {
		if len(*p) == 0 {
			break
		}
		llb.pushBack(&node{buf: (*p)[i]})
	}
}

var WriteCountErr = errors.New("LinkListBuffer.WriteTo: invalid Write count")
var ReadCountErr = errors.New("LinkListBuffer.ReadFrom: invalid WriteToSlice count")

// WriteTo implements io.WriterTo.
func (llb *LinkListBuffer) WriteTo(w io.Writer) (n int64, err error) {
	var m int
	for b := llb.pop(); b != nil; b = llb.pop() {
		m, err = w.Write(b.buf)
		if m > b.len() {
			return int64(m), WriteCountErr
		}
		n += int64(m)
		if err != nil {
			return
		}
		if m < b.len() {
			discardHead(&b.buf, m)
			llb.pushFront(b)
			return n, io.ErrShortWrite
		}
		zpool.Put(b.buf)
	}
	return
}

// ReadFrom implements io.ReaderFrom.
func (llb *LinkListBuffer) ReadFrom(r io.Reader) (int64, error) {
	b := zpool.Get()
	n, err := r.Read(b)
	if n < 0 {
		return int64(n), ReadCountErr
	}
	b = b[:n]

	if err != nil && err != io.EOF {
		zpool.Put(b)
		return int64(n), err
	}
	llb.pushBack(&node{buf: b})
	return int64(n), nil
}

// pushFront adds the new node to the head of l.
func (llb *LinkListBuffer) pushFront(b *node) {
	if b == nil {
		return
	}
	if llb.head == nil {
		b.next = nil
		llb.tail = b
	} else {
		b.next = llb.head
	}
	llb.head = b
	llb.size++
	llb.bytes += b.len()
}

// pushBack adds a new node to the tail of l.
func (llb *LinkListBuffer) pushBack(b *node) {
	if b == nil {
		return
	}
	if llb.tail == nil {
		llb.head = b
	} else {
		llb.tail.next = b
	}
	b.next = nil
	llb.tail = b
	llb.size++
	llb.bytes += b.len()
}

// pop returns and removes the head of l. If l is empty, it returns nil.
func (llb *LinkListBuffer) pop() *node {
	if llb.head == nil {
		return nil
	}
	b := llb.head
	llb.head = b.next
	if llb.head == nil {
		llb.tail = nil
	}
	b.next = nil
	llb.size--
	llb.bytes -= b.len()
	return b
}
