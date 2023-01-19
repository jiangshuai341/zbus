package reactor

import (
	"github.com/jiangshuai341/zbus/zbuffer"
	"github.com/jiangshuai341/zbus/znet/socket"
	"github.com/jiangshuai341/zbus/znet/tcp-linux/epoll"
	"net"
	"os"
	"syscall"
)

type INetHandle interface {
	OnTraffic(*zbuffer.CombinesBuffer)
	OnClose()
}

type Connection struct {
	fd             int                     // file descriptor
	localAddr      net.Addr                // local addr
	remoteAddr     net.Addr                // remote addr
	outboundBuffer *zbuffer.LinkListBuffer // 出栈缓冲区
	inboundBuffer  *zbuffer.CombinesBuffer // 入栈缓冲区
	reactor        *Reactor
	INetHandle
}

func newTCPConn(fd int) (*Connection, error) {
	if err := os.NewSyscallError("fcntl nonblock", syscall.SetNonblock(fd, true)); err != nil {
		return nil, err
	}
	lsa, err := syscall.Getsockname(fd)
	if err != nil {
		return nil, err
	}
	rsa, err := syscall.Getpeername(fd)
	if err != nil {
		return nil, err
	}

	return &Connection{
		fd:             fd,
		localAddr:      socket.SockaddrToTCPOrUnixAddr(lsa),
		remoteAddr:     socket.SockaddrToTCPOrUnixAddr(rsa),
		outboundBuffer: zbuffer.NewLinkListBuffer(),
		inboundBuffer:  zbuffer.NewCombinesBuffer(1024 * 2),
	}, nil
}

// SendSafeZeroCopy 线程安全
func (c *Connection) SendSafeZeroCopy(data ...[]byte) error {
	return c.reactor.epoller.AppendTask(func(p *epoll.Epoller) {
		c.SendUnsafeZeroCopy(data...)
	})
}

// SendSafeZeroCopy 非线程安全
func (c *Connection) SendUnsafeZeroCopy(data ...[]byte) {
	c.write(data...)
}
func (c *Connection) onRemoteClose() {
	delete(c.reactor.conns, c.fd)
	_ = c.reactor.epoller.Delete(c.fd)
	c.INetHandle.OnClose()
}

func (c *Connection) write(data ...[]byte) {
	if c.outboundBuffer.IsEmpty() {
		c.outboundBuffer.PushsNoCopy(&data)
		//writeSocketDirectly 利用EPOLLET的虹吸效应 激活EPOLL ET
		c.onTriggerWrite()
	} else {
		c.outboundBuffer.PushsNoCopy(&data)
	}
}

func (c *Connection) onTraffic() {
	for {
		c.reactor.riovc.SetPrefix(c.inboundBuffer.PeekRingBufferFreeSpace())
		n, err := epoll.Readv(c.fd, c.reactor.riovc.BufferWithPrefix())
		if err == syscall.EAGAIN || err == syscall.EINTR || n == 0 {
			break
		}
		if n < 0 || err != nil {
			c.onRemoteClose()
			log.Errorf("[onTraffic] [Connection will close] syscall Readv return:%d err:%+v ", n, err)
			return
		}
		n -= c.inboundBuffer.UpdateDataSpaceNum(n)
		c.inboundBuffer.PushsNoCopy(c.reactor.riovc.MoveTemp(n))
	}
	c.INetHandle.OnTraffic(c.inboundBuffer)
}

func (c *Connection) onTriggerWrite() {
	if c.outboundBuffer.ByteLength() == 0 {
		return
	}
	for {
		c.outboundBuffer.PeekToIovecs(&c.reactor.wiovc)
		n, err := epoll.Writev(c.fd, c.reactor.wiovc)
		if err == syscall.EAGAIN || err == syscall.EINTR || n == 0 {
			c.outboundBuffer.Discard(n)
			break
		}
		if n < 0 || err != nil {
			c.onRemoteClose()
			log.Errorf("[onTriggerWrite] [Connection will close] syscall Writev return:%d err:%+v ", n, err)
			return
		}
		c.outboundBuffer.Discard(n)
	}
}
