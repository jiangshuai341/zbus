package slicepool

var defaultBufferPool = slicePool[byte]{
	defaultBitSize: minBitSize,
}

func GetBuffer() []byte          { return defaultBufferPool.Get() }
func GetBuffer2(size int) []byte { return defaultBufferPool.Get2(size) }
func PutBuffer(b []byte)         { defaultBufferPool.Put(b) }
