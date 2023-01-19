//go:build linux

package slicepool

import "github.com/jiangshuai341/zbus/znet/tcp-linux/epoll"

var defaultIovcPool = slicePool[epoll.Iovec]{
	defaultBitSize: minBitSize,
}

func GetIovc() []epoll.Iovec          { return defaultIovcPool.Get() }
func GetIovc2(size int) []epoll.Iovec { return defaultIovcPool.Get2(size) }
func PutIovc(b []epoll.Iovec)         { defaultIovcPool.Put(b) }
