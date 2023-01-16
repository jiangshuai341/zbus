package reactor

import (
	"github.com/jiangshuai341/zbus/znet/epoll"
	"github.com/jiangshuai341/zbus/znet/socket"
	"runtime"
	"syscall"
)

type Accepter struct {
	ep  *epoll.Epoller
	cb  func(conn *Connection)
	lfd []int
}

//当Accept成为系统瓶颈时，建议使用端口复用，多线程Accept （HTTP短连接服务）
//当有多个端口需要Accept,并不构成系统瓶颈时可以聚合到一个Epoller进行Accept （TCP长连接服务）

// ActiveListener 将传入的ListenSocket 通过IO复用同时监听 会启动一个LockOSThread的IO线程
func ActiveListener(addr string, number int, OnAccept func(conn *Connection)) (ret []*Accepter, err error) {
	var fd int
	var ep *epoll.Epoller

	for i := 0; i < number; i++ {

		if fd, _, err = socket.TCPSocket(socket.TCP, addr, true, socket.Option{SetSockOpt: socket.SetReuse}); err != nil {
			break
		}

		if ep, err = epoll.OpenEpoller(); err != nil {
			break
		}

		accepter := &Accepter{
			ep: ep,
			cb: OnAccept,
		}

		if err = accepter.AddListen(fd); err != nil {
			_ = accepter.ep.Close()
			break
		}

		go func() {
			runtime.LockOSThread()
			pollingErr := accepter.ep.Epolling(accepter.onAccept)
			if pollingErr != nil {
				log.Error(pollingErr.Error())
			}
		}()

		ret = append(ret, accepter)
	}

	return ret, err
}

/*
C++ EpollET Accept例子
while ((conn_sock = accept(listenfd,(struct sockaddr *) &remote, (size_t *)&addrlen)) > 0)
{
    handle_client(conn_sock);
}
if (conn_sock == -1)
{
    if (errno != EAGAIN && errno != ECONNABORTED && errno != EPROTO && errno != EINTR)
    perror("accept");
}
*/

// onAccept 执行线程 IO Thread
func (a *Accepter) onAccept(lfd int, _ uint32) {
	var fd int
	var err error
	var sa syscall.Sockaddr
	for {
		fd, sa, err = syscall.Accept(lfd)
		if fd <= 0 {
			break
		}
		conn, err2 := newTCPConn(fd)
		if err2 != nil {
			log.Errorf("[Accepter] fd:%d RemoteAddr:%+v Err:%s", fd, sa, err2.Error())
			continue
		}
		a.cb(conn)
		if conn.INetHandle == nil {
			panic("[Accepter] Please Check Code And Implement Connection.INetHandle")
		}
	}
	if fd == -1 &&
		err != syscall.EAGAIN &&
		err != syscall.ECONNABORTED &&
		err != syscall.EPROTO &&
		err != syscall.EINTR {
		log.Errorf("[Accepter] syscall.Accept Failed ListenFD:%d  Err:%s", a.lfd, err.Error())
		a.Close()
	}
	return
}

// AddListen 线程安全
func (a *Accepter) AddListen(fd int) error {
	return a.ep.AppendUrgentTask(func(p *epoll.Epoller) {
		err := p.AddRead(fd)
		if err != nil {
			log.Errorf("Epoller AddRead Failed Socketname:%+v", fd, any(syscall.Getsockname(fd)))
			_ = syscall.Close(fd)
			return
		}
		a.lfd = append(a.lfd, fd)
	})
}

func (a *Accepter) Close() {
	_ = a.ep.AppendUrgentTask(func(e *epoll.Epoller) {
		for _, v := range a.lfd {
			_ = syscall.Close(v)
		}
		_ = e.Close()
	})
}
