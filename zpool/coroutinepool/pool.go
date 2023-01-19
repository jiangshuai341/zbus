package coroutinepool

import "hash/crc32"

func submit(hashKey []byte, fun func()) {
	crc32.ChecksumIEEE(hashKey)
}
