package socket

import (
	"errors"
	"os"
	"syscall"
)

// Option is used for setting an option on socket.
type Option struct {
	SetSockOpt func(int, int) error
	Opt        int
}

//SO_REUSEADDR 之后可以多路bind 但是处于listen状态不可以
//SO_REUSEPORT 之后可以多路监听

func SetReuse(fd int, _ int) error {
	const SO_REUSEPORT = 0xf
	// SetReuseAddr enables SO_REUSEADDR option on socket.
	err := os.NewSyscallError("Set Socket Reuse", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1))
	if err != nil {
		return err
	}
	// SetReuseport enables SO_REUSEPORT option on socket.
	err = os.NewSyscallError("Set Socket Reuse", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, SO_REUSEPORT, 1))
	if err != nil {
		return err
	}
	return nil
}

func SetNoDelay(fd, _ int) error {
	return os.NewSyscallError("Set Socket Close Nagle's algorithm", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1))
}

// SetTcpKeepIntvl 心跳探测间隔 TCP_KEEPINTVL 覆盖 tcp_keepalive_intvl，默认75（秒）
func SetTcpKeepIntvl(fd, secs int) error {
	if secs <= 0 {
		return errors.New("invalid time duration")
	}
	return os.NewSyscallError("SetTcpKeepIntvl", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, secs))
}

// SetTcpKeepCnt n次无响应关闭对端 TCP_KEEPCNT 覆盖 tcp_keepalive_probes，默认9（次）
func SetTcpKeepCnt(fd, n int) error {
	if n <= 0 {
		return errors.New("invalid time duration")
	}
	return os.NewSyscallError("SetTcpKeepCnt", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, n))
}

// SetTcpKeepIdle 无数据交互secs秒之后开始心跳检测 TCP_KEEPIDLE 覆盖 tcp_keepalive_time，默认7200（秒）
func SetTcpKeepIdle(fd, secs int) error {
	if secs <= 0 {
		return errors.New("invalid time duration")
	}
	return os.NewSyscallError("SetTcpKeepIdle", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, secs))
}

// SetTcpKeepAlive 开启TCP OS 级别心跳检测
func SetTcpKeepAlive(fd, _ int) error {
	return os.NewSyscallError("SetTcpKeepAlive", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1))
}

//非丢包延时大的情况下可以适当调大缓冲区
//假设带宽为1000 Mbit/s，rtt时间为400ms，那么缓存应该调整为大约50Mbyte左右

// SetRecvBuffer 设置Socket接收缓冲区大小
func SetRecvBuffer(fd, size int) error {

	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, size)
}

// SetSendBuffer 设置Socket发送缓冲区大小 UDP没有发送缓冲区直接offload到网卡
func SetSendBuffer(fd, size int) error {
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_SNDBUF, size)
}
