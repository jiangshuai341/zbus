package znet

import (
	"github.com/jiangshuai341/zbus/znet/reactor"
	"net"
	"testing"
	"time"
)

type TcpTask struct {
}

func (t *TcpTask) OnTraffic(data *[][]byte) bool {

	return true
}
func (t *TcpTask) OnClose() {

}

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
func OnAccept(conn *reactor.Connection) {
	conn.INetHandle = &TcpTask{}
	err := reactorMgr.LoadBalance().AddConn(conn)
	if err != nil {
		return
	}
}

var accepters = make([]*reactor.Accepter, 10)

func TestListen(t *testing.T) {
	var err error
	accepters, err = reactor.ActiveListener("0.0.0.0:9999", 1, OnAccept)
	if err != nil {
		t.Logf("ActiveListener Failed Err:%+v", err)
		return
	}
	time.Sleep(1000000 * time.Second)
}

func TestClient(t *testing.T) {
	net.Dial("tcp", "0.0.0.0:9999")
	time.Sleep(1000000 * time.Second)
}

func BenchmarkSend(b *testing.B) {
	for i := 0; i < b.N; i++ {

	}
	b.Log()
}

func BenchmarkRecv(b *testing.B) {
	for i := 0; i < b.N; i++ {

	}
	b.Log()
}
