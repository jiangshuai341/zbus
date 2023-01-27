package service

import (
	"github.com/jiangshuai341/zbus/znet/tcp-linux/reactor"
)

// 服务器集群通信模块

type network struct {
	r        *reactor.Reactor
	a        *reactor.Accepter
	onAccept func(conn *reactor.Connection)
}

func (s *Service) GetConn(serviceName string, serviceId int32) {

}
func (s *Service) networkStart() {
	var err error
	s.r, err = reactor.NewReactor()
	s.a, err = reactor.NewListener(s.onAccept)
	for _, url := range s.serviceUrl {
		err = s.a.ListenUrl(url)
		if err != nil {
			panic(s.etcdkey() + s.etcdval() + "listen url failed" + url)
		}
	}
}

//func (s *Service) onAccept(conn *reactor.Connection) {
//	conn.INetHandle = &netTask{c: conn}
//	err := s.r.AddConn(conn)
//	if err != nil {
//		return
//	}
//}

//type netTask struct {
//	c *reactor.Connection
//}
//
//func (e *netTask) OnTraffic(inboundBuffer *zbuffer.CombinesBuffer) {
//	pakSize, err := inboundBuffer.PeekInt(0, 4)
//
//	if err != nil {
//		return
//	}
//	dataLen := inboundBuffer.LengthData()
//	if int(pakSize) > dataLen-4 {
//		return
//	}
//
//	cmd, err := inboundBuffer.PeekInt(4, 4)
//	if err != nil {
//		return
//	}
//
//	switch cmd {
//
//	}
//
//	//t.c.SendUnsafeNoCopy(*inboundBuffer.PopData(int(pakSize + 4)))
//}
//
//func (e *netTask) OnClose() {
//
//}
