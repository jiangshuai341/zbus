package znet

import (
	"github.com/jiangshuai341/zbus/logger"
	"github.com/jiangshuai341/zbus/toolkit"
	"github.com/jiangshuai341/zbus/zbuf"
	"github.com/jiangshuai341/zbus/znet/reactor"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
	"unsafe"
)

type TcpTask struct {
	c *reactor.Connection
}

var testLog = logger.GetLogger("ReactorTest")

func (t *TcpTask) OnTraffic(inboundBuffer *zbuf.CombinesBuffer) {
	pakSize, err := inboundBuffer.PeekInt(4)
	if err != nil {
		return
	}
	dataLen := inboundBuffer.LengthData()

	if int(pakSize) > dataLen-4 {
		return
	}

	inboundBuffer.Discard(4)

	inboundBuffer.PopData(int(pakSize))
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
	conn.INetHandle = &TcpTask{c: conn}
	err := reactorMgr.LoadBalance().AddConn(conn)
	if err != nil {
		return
	}
}

//var accepters = make([]*reactor.Accepter, 10)

func TestListen(t *testing.T) {
	var err error
	_, err = reactor.ActiveListener("0.0.0.0:9999", 1, OnAccept)
	if err != nil {
		t.Logf("ActiveListener Failed Err:%+v", err)
		return
	}
	time.Sleep(1000000 * time.Second)
}

func TestClient(t *testing.T) {
	conn, _ := net.Dial("tcp", "0.0.0.0:9999")
	var num = 1
	var syncCtx sync.WaitGroup
	syncCtx.Add(1)
	go func() {
		var tempRead []byte = make([]byte, 1024)

		var tempWrite []byte
		a := (*reflect.SliceHeader)(unsafe.Pointer(&tempWrite))
		a.Len = 8
		a.Data = uintptr(unsafe.Pointer(&num))
		a.Cap = 8

		var readNum *int64 = (*int64)(unsafe.Pointer(&tempRead[0]))

		for {
			n, err := conn.Write(tempWrite)
			if err != nil {
				break
			}
			testLog.Infof("num:%d ret:%d err:%+v", n, num, err)
			n, err = conn.Read(tempRead)
			if err != nil {
				break
			}
			testLog.Infof("num:%d ret:%d err:%+v", n, *readNum, err)
			num++
		}
		syncCtx.Done()
	}()

	syncCtx.Wait()
}

func BenchmarkSend(b *testing.B) {
	conn, err := net.Dial("tcp", "0.0.0.0:9999")
	if err != nil {
		b.Log(err)
	}
	var str string = "hhhhhhhh"
	for i := 0; i < b.N; i++ {
		_, err = conn.Write(toolkit.StringToBytes(str))
		if err != nil {
			b.Log(err)
		}
	}
}

func BenchmarkRecv(b *testing.B) {
	for i := 0; i < b.N; i++ {

	}

	b.Log()
}
