package znet

import (
	"github.com/jiangshuai341/zbus/zbuf"
	"github.com/jiangshuai341/zbus/znet/reactor"
	"io"
	"net"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func TestListen(t *testing.T) {

	time.Sleep(1000000 * time.Second)
}

func TestClient(t *testing.T) {

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

func runServer(num int) {
	_, _ = reactor.ActiveListener("0.0.0.0:9999", num, OnAccept)
}

func runClient(dataBlockSize int, clientNum int) {
	for i := 0; i < clientNum; i++ {
		conn, _ := net.Dial("tcp", "0.0.0.0:9999")
		var syncCtx sync.WaitGroup
		syncCtx.Add(1)

		var pakData = func(data []byte) []byte {
			var ret = make([]byte, 4, 4+len(data))
			*(*int32)(unsafe.Pointer(&ret[0])) = int32(len(data))
			ret = append(ret, data...)
			return ret
		}

		go func() {
			var tempRead = make([]byte, dataBlockSize+4)
			var tempWrite = make([]byte, dataBlockSize)
			for {
				_, err := conn.Write(pakData(tempWrite))
				if err != nil {
					break
				}
				_, err = io.ReadAtLeast(conn, tempRead, len(tempRead))
				if err != nil {
					break
				}
			}
		}()
	}
}

type NetTask struct {
	c *reactor.Connection
}

func (t *NetTask) OnTraffic(inboundBuffer *zbuf.CombinesBuffer) {
	pakSize, err := inboundBuffer.PeekInt(4)
	if err != nil {
		return
	}
	dataLen := inboundBuffer.LengthData()
	if int(pakSize) > dataLen-4 {
		return
	}

	t.c.SendUnsafeNoCopy(*inboundBuffer.PopData(int(pakSize + 4)))
}

func (t *NetTask) OnClose() {

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
	conn.INetHandle = &NetTask{c: conn}
	err := reactorMgr.LoadBalance().AddConn(conn)
	if err != nil {
		return
	}
}
