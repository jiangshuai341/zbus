package reactor

import (
	"errors"
	"github.com/jiangshuai341/zbus/zbuf"
	"github.com/jiangshuai341/zbus/znet/epoll"
	"github.com/jiangshuai341/zbus/znet/socket"
	"net"
	"os"
	"syscall"
)

type INetHandle interface {
	OnTraffic(data *[][]byte) bool
	OnClose()
}

type Connection struct {
	fd             int                  // file descriptor
	localAddr      net.Addr             // local addr
	remoteAddr     net.Addr             // remote addr
	outboundBuffer *zbuf.LinkListBuffer // 出栈缓冲区
	inboundBuffer  *zbuf.LinkListBuffer // 入栈缓冲区
	reactor        *Reactor
	tempPeek       [][]byte
	INetHandle
}

var errNetHandleImpIsNil = errors.New("[newTCPConn] NetHandleImp Is Nil")

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
		outboundBuffer: zbuf.NewLinkListBuffer(),
		inboundBuffer:  zbuf.NewLinkListBuffer(),
		tempPeek:       make([][]byte, 0, 8),
	}, nil
}

func (c *Connection) SendSafe(data []byte) error {
	return c.reactor.epoller.AppendTask(func(arg ...any) {
		for _, v := range arg {
			c.SendUnsafe(v.([]byte))
		}
	}, data)
}

//SendSafeNoCopy 线程安全
func (c *Connection) SendSafeNoCopy(data []byte) error {
	return c.reactor.epoller.AppendTask(func(arg ...any) {
		for _, v := range arg {
			c.SendUnsafeNoCopy(v.([]byte))
		}
	}, data)
}

//SendUnsafe 非线程安全
func (c *Connection) SendUnsafe(data []byte) {
	temp := c.outboundBuffer.NewBytesFromPool(len(data))
	copy(temp, data)
	c.write(temp)
}

func (c *Connection) SendUnsafeNoCopy(data []byte) {
	c.write(data)
}
func (c *Connection) onRemoteClose() {
	delete(c.reactor.conns, c.fd)
	_ = c.reactor.epoller.Delete(c.fd)
	c.INetHandle.OnClose()
}

func (c *Connection) write(data []byte) {
	if c.outboundBuffer.IsEmpty() {
		c.outboundBuffer.PushNoCopy(&data)
		//writeSocketDirectly 利用EPOLLET的虹吸效应 激活EPOLL ET
		c.onTriggerWrite()
	} else {
		c.outboundBuffer.PushNoCopy(&data)
	}
}

func (c *Connection) onTraffic() {
	for {
		//temp := make([]byte, 1000)
		//n, err := syscall.Read(c.fd, temp)
		n, err := epoll.Readv(c.fd, *c.reactor.tempReadBuffer.Buffer())

		//syscall.EWOULDBLOCK == syscall.EAGAIN
		if err == syscall.EAGAIN || err == syscall.EINTR || n == 0 {
			c.inboundBuffer.PushsNoCopy(c.reactor.tempReadBuffer.MoveTemp(n))
			break
		}
		if n < 0 || err != nil {
			c.onRemoteClose()
			log.Errorf("[onTraffic] [Connection will close] syscall Readv return:%d err:%+v ", n, err)
			return
		}
		c.inboundBuffer.PushsNoCopy(c.reactor.tempReadBuffer.MoveTemp(n))
	}
	c.tempPeek = c.tempPeek[:0]
	c.inboundBuffer.Peek(-1, &c.tempPeek)
	c.INetHandle.OnTraffic(&c.tempPeek)
}

func (c *Connection) onTriggerWrite() {
	if c.outboundBuffer.ByteLength() == 0 {
		return
	}
	for {
		c.outboundBuffer.Peek(-1, &c.reactor.tempWriteBuffer)
		n, err := epoll.Writev(c.fd, c.reactor.tempWriteBuffer)
		if err == syscall.EAGAIN || err == syscall.EINTR || n == 0 {
			c.outboundBuffer.DiscardBytes(n)
			break
		}
		if n < 0 || err != nil {
			c.onRemoteClose()
			log.Errorf("[onTriggerWrite] [Connection will close] syscall Writev return:%d err:%+v ", n, err)
			return
		}
		c.outboundBuffer.DiscardBytes(n)
	}
}
