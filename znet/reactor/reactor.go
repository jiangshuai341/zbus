package reactor

import (
	"errors"
	"github.com/jiangshuai341/zbus/logger"
	"github.com/jiangshuai341/zbus/zbuf"
	"github.com/jiangshuai341/zbus/znet/epoll"
	"github.com/jiangshuai341/zbus/znet/socket"
	"syscall"
)

var log = logger.GetLogger("reactor")
var ErrNetHandle = errors.New("please init INetHandle conn before add")

type Reactor struct {
	epoller         *epoll.Epoller
	conns           map[int]*connection
	tempReadBuffer  zbuf.ArrayBuffer
	tempWriteBuffer [][]byte
}

func NewReactor() (r *Reactor, err error) {
	r = &Reactor{conns: make(map[int]*connection)}
	r.tempReadBuffer.Reserve(socket.SocketReadBufferSize)
	r.epoller, err = epoll.OpenEpoller()
	if err != nil {
		return nil, err
	}
	go func() {
		epollErr := r.epoller.Epolling(r.OnReadWriteEventTrigger)
		if err != nil {
			log.Errorf("reactor epoll systemcall err:%+v , quit reactor", epollErr)
		}
	}()
	return
}

//DoTaskInIoThread
func (r *Reactor) DoTaskInIoThread(fn epoll.TaskFunc, arg ...any) error {
	return r.epoller.AppendTask(fn, arg)
}

//OnReadWriteEventTrigger Trigger On Io Thread
func (r *Reactor) OnReadWriteEventTrigger(fd int, ev uint32) {
	conn, ok := r.conns[fd]
	if !ok {
		log.Errorf("please check <add,remove> ds:%d is not exist in conns map", fd)
		return
	}
	if ev&syscall.EPOLLRDHUP != 0 { //对端关闭
		conn.onRemoteClose()
	}
	if ev&syscall.EPOLLOUT != 0 { //Socket发送缓冲区状态 写满 -> 可写
		conn.onTriggerWrite()
	}
	if ev&syscall.EPOLLIN != 0 { //Socket接收缓冲区状态 空 -> 可读
		conn.onTraffic()
	}
}

//AddConn
func (r *Reactor) AddConn(conn *connection) error {
	if conn.INetHandle == nil {
		return ErrNetHandle
	}
	r.conns[conn.fd] = conn
	return r.epoller.AddReadWrite(conn.fd)
}
