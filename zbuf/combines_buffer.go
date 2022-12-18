package zbuf

type CombinesBuffer struct {
	ringBuffer *RingBuffer
	listBuffer *LinkListBuffer
}

func NewCombinesBuffer(ringSize int) *CombinesBuffer {
	return &CombinesBuffer{
		ringBuffer: NewRingBuffer(ringSize),
		listBuffer: NewLinkListBuffer(),
	}
}
func (c *CombinesBuffer) PeekFreeSpace() (head []byte, tail []byte) {
	if !c.listBuffer.IsEmpty() {
		return nil, nil
	}
	return c.ringBuffer.PeekFreeSpace()
}

func (c *CombinesBuffer) PeekDataSpace(ret *[][]byte) {
	head, tail := c.ringBuffer.PeekDataSpace()
	if head == nil {
	} else if tail == nil {
		*ret = append(*ret, head)
	} else {
		*ret = append(*ret, head)
		*ret = append(*ret, tail)
	}
	if !c.listBuffer.IsEmpty() {
		c.listBuffer.Peek(-1, ret)
	}
	return
}
func (c *CombinesBuffer) Discard(num int) int {
	temp := num
	temp -= c.ringBuffer.Discard(num)
	if num > 0 {
		temp -= c.listBuffer.DiscardBytes(num)
	}
	return num - temp
}
func (c *CombinesBuffer) UpdateDataSpaceNum(newWriteNum int) int {
	return c.ringBuffer.DataSpaceGrow(newWriteNum)
}
func (c *CombinesBuffer) LengthData() int {
	return c.ringBuffer.LengthData() + c.listBuffer.ByteLength()
}

func (c *CombinesBuffer) PushsNoCopy(temp *[][]byte) {
	c.listBuffer.PushsNoCopy(temp)
}
