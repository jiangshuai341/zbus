package znet

import (
	"github.com/jiangshuai341/zbus/zbuffer"
	"github.com/jiangshuai341/zbus/znet/linux_tcp/reactor"
)

var reactorMgr = NewReactorMgr()

type ReactorMgr struct {
	reactors []*reactor.Reactor
}

type NetTask struct {
	c *reactor.Connection
}

func (e *ReactorMgr) LoadBalance() *reactor.Reactor {
	return e.reactors[0]
}
func NewReactorMgr() (e *ReactorMgr) {
	e = &ReactorMgr{reactors: make([]*reactor.Reactor, 10)}
	for i := 0; i < len(e.reactors); i++ {
		e.reactors[i], _ = reactor.NewReactor()
	}
	return
}

func OnAccept(conn *reactor.Connection) {
	conn.INetHandle = &NetTask{c: conn}
	err := reactorMgr.LoadBalance().AddConn(conn)
	if err != nil {
		return
	}
}

func (t *NetTask) OnTraffic(inboundBuffer *zbuffer.CombinesBuffer) {
	pakSize, err := inboundBuffer.PeekInt(0, 4)
	if err != nil {
		return
	}
	dataLen := inboundBuffer.LengthData()
	if int(pakSize) > dataLen-4 {
		return
	}

	t.c.SendUnsafeZeroCopy(inboundBuffer.PopsData(int(pakSize + 4))...)
}

func (t *NetTask) OnClose() {

}
