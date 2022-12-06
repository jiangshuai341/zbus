package reactor

import (
	"github.com/jiangshuai341/zbus/zbuf"
	"github.com/jiangshuai341/zbus/znet/epoll"
	"net"
	"syscall"
)

type INetHandle interface {
	OnTraffic(data *[][]byte) bool
	OnClose()
}

const MaxBufferLength int32 = 10240 //bytes

type connection struct {
	localAddr      net.Addr             // local addr
	remoteAddr     net.Addr             // remote addr
	outboundBuffer *zbuf.LinkListBuffer // 出栈缓冲区
	inboundBuffer  *zbuf.LinkListBuffer // 入栈缓冲区
	buffer         []byte               // buffer for the latest bytes
	fd             int                  // file descriptor
	reactor        *Reactor
	tempPeek       [][]byte
	isActiveWrite  bool
	INetHandle
}

func (c *connection) SendSafe() {

}

//SendSafeNoCopy 在IO线程中调用
func (c *connection) SendSafeNoCopy(data []byte) error {
	return c.reactor.epoller.AppendTask(func(arg ...any) {

	}, data)
}

//SendUnsafe 必须在IO线程中调用
func (c *connection) SendUnsafe(data []byte) error {
	return nil
}

func (c *connection) SendUnsafeNoCopy(data []byte) error {
	return nil
}
func (c *connection) onRemoteClose() {
	delete(c.reactor.conns, c.fd)
	_ = c.reactor.epoller.Delete(c.fd)
	c.INetHandle.OnClose()
}

func (c *connection) write(data []byte) {
	if c.outboundBuffer.IsEmpty() {
		c.outboundBuffer.PushNoCopy(&data)
		//writeSocketDirectly 利用EPOLLET的虹吸效应 激活EPOLL ET
		c.onTriggerWrite()
	} else {
		c.outboundBuffer.PushNoCopy(&data)
	}
}

func (c *connection) onTraffic() {
	for {
		n, err := epoll.Readv(c.fd, *c.reactor.tempReadBuffer.Buffer())
		//syscall.EWOULDBLOCK == syscall.EAGAIN
		if err == syscall.EAGAIN || err == syscall.EINTR {
			c.inboundBuffer.PushsNoCopy(c.reactor.tempReadBuffer.MoveTemp(n))
			break
		}
		if n < 0 || err != nil {
			c.onRemoteClose()
			log.Errorf("[onTraffic] [connection will close] syscall Readv return:%d err:%+v ", n, err)
			return
		}
		c.inboundBuffer.PushsNoCopy(c.reactor.tempReadBuffer.MoveTemp(n))
	}
	c.inboundBuffer.Peek(-1, &c.tempPeek)
	c.INetHandle.OnTraffic(&c.tempPeek)
}

func (c *connection) onTriggerWrite() {
	for {
		c.outboundBuffer.Peek(-1, &c.reactor.tempWriteBuffer)
		n, err := epoll.Writev(c.fd, c.reactor.tempWriteBuffer)
		if err == syscall.EAGAIN || err == syscall.EINTR {
			c.outboundBuffer.DiscardBytes(n)
			break
		}
		if n < 0 || err != nil {
			c.onRemoteClose()
			log.Errorf("[onTriggerWrite] [connection will close] syscall Writev return:%d err:%+v ", n, err)
			return
		}
		c.outboundBuffer.DiscardBytes(n)
	}
}
