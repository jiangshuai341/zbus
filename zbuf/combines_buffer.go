package zbuf

import (
	"errors"
	"github.com/jiangshuai341/zbus/zpool"
	"math"
	"unsafe"
)

type CombinesBuffer struct {
	ringBuffer *RingBuffer
	listBuffer *LinkListBuffer
	peekTemp   [][]byte
}

func NewCombinesBuffer(ringSize int) *CombinesBuffer {
	return &CombinesBuffer{
		ringBuffer: NewRingBuffer(ringSize),
		listBuffer: NewLinkListBuffer(),
		peekTemp:   make([][]byte, 0),
	}
}
func (c *CombinesBuffer) PeekFreeAll() (head []byte, tail []byte) {
	if !c.listBuffer.IsEmpty() {
		return nil, nil
	}
	return c.ringBuffer.PeekFreeSpace()
}

func (c *CombinesBuffer) PeekDataAll() *[][]byte {
	c.peekTemp = c.peekTemp[:0]
	head, tail := c.ringBuffer.PeekDataSpace()
	if head == nil {
	} else if tail == nil {
		c.peekTemp = append(c.peekTemp, head)
	} else {
		c.peekTemp = append(c.peekTemp, head)
		c.peekTemp = append(c.peekTemp, tail)
	}
	if !c.listBuffer.IsEmpty() {
		c.listBuffer.Peek(-1, &c.peekTemp)
	}
	return &c.peekTemp
}

func (c *CombinesBuffer) PeekData(num int) *[][]byte {
	if num < 0 {
		return c.PeekDataAll()
	}

	c.peekTemp = c.peekTemp[:0]
	head, tail := c.ringBuffer.PeekDataSpace()

	if head != nil {
		c.peekTemp = append(c.peekTemp, head)
		num -= len(head)
	}
	if num <= 0 {
		return &c.peekTemp
	}

	if tail != nil {
		c.peekTemp = append(c.peekTemp, tail)
		num -= len(tail)
	}

	if num <= 0 {
		return &c.peekTemp
	}

	if !c.listBuffer.IsEmpty() {
		c.listBuffer.Peek(num, &c.peekTemp)
	}

	return &c.peekTemp
}

func (c *CombinesBuffer) PopData(num int) *[]byte {
	if num > c.LengthData() {
		return nil
	}
	if num < 0 {
		num = c.LengthData()
	}
	ret := zpool.Get2(num)
	var n int
	for _, v := range *c.PeekData(num) {
		n += copy(ret[n:], v)
	}
	c.Discard(num)
	return &ret
}

func (c *CombinesBuffer) Discard(num int) int {
	temp := num
	temp -= c.ringBuffer.Discard(num)
	if temp > 0 {
		temp -= c.listBuffer.DiscardBytes(temp)
	}
	return num - temp
}

// UpdateDataSpaceNum 返回有多少数据成功标记写入
func (c *CombinesBuffer) UpdateDataSpaceNum(newWriteNum int) int {
	return c.ringBuffer.DataSpaceGrow(newWriteNum)
}
func (c *CombinesBuffer) LengthData() int {
	return c.ringBuffer.LengthData() + c.listBuffer.ByteLength()
}

func (c *CombinesBuffer) PushsNoCopy(temp *[][]byte) {
	c.listBuffer.PushsNoCopy(temp)
}

var ErrDataNotEnough = errors.New("err : Data Not Enough Peek")

// PeekInt 返回读取到的整型数值 最大64位  仅小端试用
func (c *CombinesBuffer) PeekInt(byteNum int) (uint64, error) {
	if c.LengthData() < byteNum {
		return math.MaxUint64, ErrDataNotEnough
	}
	if byteNum > 8 || byteNum < 0 {
		byteNum = 8
	}
	var tempBytes []byte = make([]byte, 8)
	var n = 0
	for _, v := range *c.PeekData(byteNum) {
		n += copy(tempBytes[n:], v)
		if n >= byteNum {
			break
		}
	}
	return *(*uint64)(unsafe.Pointer(&tempBytes[0])), nil
}
