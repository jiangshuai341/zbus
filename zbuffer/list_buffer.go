package zbuffer

import (
	"github.com/jiangshuai341/zbus/zpool"
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
	return zpool.GetBuffer2(len)
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
		zpool.PutBuffer(b.buf)
	}
	llb.head = nil
	llb.tail = nil
	llb.size = 0
	llb.bytes = 0
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
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

// Discard removes some nodes based on n bytes.
func (llb *LinkListBuffer) Discard(n int) (discarded int) {
	if n <= 0 {
		return
	}
	for n != 0 {
		b := llb.pop()
		if b == nil {
			break
		}
		if n < b.len() {
			b.buf = b.buf[n:]
			discarded += n
			llb.pushFront(b)
			break
		}
		n -= b.len()
		discarded += b.len()
		zpool.PutBuffer(b.buf)
	}
	return
}

func (llb *LinkListBuffer) Push(p []byte) {
	n := len(p)
	if n == 0 {
		return
	}
	b := zpool.GetBuffer2(n)
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
