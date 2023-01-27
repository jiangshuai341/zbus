package znet

import (
	"github.com/jiangshuai341/zbus/zbuffer"
	"github.com/jiangshuai341/zbus/znet/tcp-linux/reactor"
)

var reactorMgr = NewReactorMgr()

type ReactorMgr struct {
	reactors []*reactor.Reactor
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

type NetTask struct {
	c *reactor.Connection
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

func runReactorServer(num int) {
	accepter, _ := reactor.NewListener(func(conn *reactor.Connection) {
		conn.INetHandle = &NetTask{c: conn}
		err := reactorMgr.LoadBalance().AddConn(conn)
		if err != nil {
			return
		}
	})
	_ = accepter.ListenUrl("tcp://0.0.0.0:9999")
}
