package reactor

import (
	"github.com/jiangshuai341/zbus/znet/epoll"
	"github.com/jiangshuai341/zbus/znet/socket"
	"net"
	"runtime"
	"syscall"
)

type Accepter struct {
	ep         *epoll.Epoller
	cb         func(nfd int)
	listenAddr net.Addr
}

//当Accept成为系统瓶颈时，建议使用端口复用，多线程Accept （HTTP短连接服务）
//当有多个端口需要Accept,并不构成系统瓶颈时可以聚合到一个Epoller进行Accept （TCP长连接服务）

// ActiveListener 将传入的ListenSocket 通过IO复用同时监听 会启动一个LockOSThread的IO线程
func ActiveListener(addr string, number int, callback func(nfd int)) (ret []*Accepter, err error) {
	var fd int
	var listenAddr net.Addr
	var ep *epoll.Epoller
	for i := 1; i < number; i++ {
		fd, listenAddr, err = socket.TCPSocket(socket.TCP, addr, true, socket.Option{SetSockOpt: socket.SetReuse})
		if err != nil {
			return nil, err
		}
		ep, err = epoll.OpenEpoller()
		if err != nil {
			return nil, err
		}
		accepter := &Accepter{
			ep:         ep,
			cb:         callback,
			listenAddr: listenAddr,
		}
		if err = accepter.AddListen(fd); err != nil {
			_ = accepter.ep.Close()
			return nil, err
		}
		go func() {
			runtime.LockOSThread()
			pollingErr := accepter.ep.Epolling(accepter.OnAccept)
			if pollingErr != nil {
				log.Error(pollingErr.Error())
			}
		}()
		ret = append(ret, accepter)
	}
	return ret, nil
}

//AddListen 线程安全
func (receiver *Accepter) AddListen(fd int) error {
	return receiver.ep.AddRead(fd)
}

//OnAccept 执行线程 IO Thread
func (receiver *Accepter) OnAccept(fd int, _ uint32) {
	for {
		nfd, sockAddr, err := syscall.Accept(fd)
		if nfd == -1 && err == syscall.EAGAIN {
			log.Debugf("[Accepter] Accept All Call Break")
			break
		} else if nfd == -1 {
			log.Errorf("[Accepter] Accept Failed Err:%s", err.Error())
			continue
		}
		log.Infof("[Accepter] Accept New FD:%d Addr:%+v", nfd, sockAddr)
	}
}
