package gproxy

import (
	"github.com/jiangshuai341/zbus/etcd"
	"github.com/jiangshuai341/zbus/toolkit"
	"github.com/jiangshuai341/zbus/zbuffer"
	"github.com/jiangshuai341/zbus/znet/tcp-linux/reactor"
	"github.com/jiangshuai341/zbus/zrpc"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"runtime"
	"strconv"
)

/*
0       2       4       6       8 (BYTE)
+-------+-------+-------+-------+
|     pakLen    |       cmd     |
+---------------+---------------+   8
|  serviceType  |    funcHash   |
+---------------+---------------+  16
|            entityID           |
+---------------+---------------+  24
|                               |
|              DATA             |
|                               |
+-------------------------------+
*/

type Server struct {
	reactor  *reactor.Reactor
	accepter *reactor.Accepter
	etcd     *etcd.Client
}

type Config struct {
	ListenPort []string `json:"ListenPort"`
	ListenUds  string   `json:"ListenUds"`
	ReportAddr []string `json:"ReportAddr"`
}

func NewServer() *Server {
	gproxy := &Server{
		reactor: NewReactor(),
	}
	var serviceId int32
	var listenAddr []string
	gproxy.accepter, serviceId, listenAddr = NewAccepter(gproxy)

	etcd.NewClient([]string{"localhost:2379"})
	gproxy.etcd = etcd.NewClient([]string{"localhost:2379"})

	gproxy.etcd.RegisterAndKeepaliveToETCD(
		"zbus",
		"gproxy",
		1,
		serviceId,
		PortMapping(listenAddr),
	)

	gproxy.etcd.WatchAllService(gproxy)
	return gproxy
}
func NewReactor() (ret *reactor.Reactor) {
	ret, err := reactor.NewReactor()
	if err != nil {
		panic("rpc proxy service new reactor failed Err:" + err.Error())
	}
	return
}
func PortMapping(listenAddr []string) (reportAddr []string) {
	return listenAddr
}
func NewAccepter(iAccepter reactor.IAccepter) (ret *reactor.Accepter, serviceId int32, listenAddr []string) {
	ret, err := reactor.NewListener(iAccepter)
	if err != nil {
		panic("rpc proxy service reactor.NewListener failed Err:" + err.Error())
	}
	port, err := toolkit.GetFreePort()
	if err != nil {
		panic("rpc proxy service toolkit.GetFreePort failed Err:" + err.Error())
	}
	if err = ret.ListenUrl("tcp://" + "0.0.0.0:" + strconv.Itoa(port)); err != nil {
		panic("rpc proxy service accepter.ListenUrl failed Err:" + err.Error())
	}
	if runtime.GOOS == "linux" {
		if err = ret.ListenUrl("unix://" + "/tmp/hhhhh.sock"); err != nil {
			panic("rpc proxy service listen unix socket failed Err:" + err.Error())
		}
	}
	return
}
func (s *Server) OnServiceStatusChange(name string, version int32, serviceId int32, urls []string, eventType mvccpb.Event_EventType) {

}

func (s *Server) OnAccept(conn *reactor.Connection) {
	conn.INetHandle = &entity{c: conn}
	err := s.reactor.AddConn(conn)
	if err != nil {
		return
	}
}

type entity struct {
	serviceType int32
	entityID    int64
	serviceMap  map[int32]string
	delegateMap map[int32]string
	c           *reactor.Connection
}

func (e *entity) OnTraffic(inboundBuffer *zbuffer.CombinesBuffer) {
	pakSize, err := inboundBuffer.PeekInt(0, 4)

	if err != nil {
		return
	}
	dataLen := inboundBuffer.LengthData()
	if int(pakSize) > dataLen-4 {
		return
	}

	cmd, err := inboundBuffer.PeekInt(4, 4)
	if err != nil {
		return
	}

	switch zrpc.Cmd(cmd) {
	case zrpc.BindDelegate:

	case zrpc.RemoteInvoke:

	case zrpc.CreateEntity:

	case zrpc.DeclareDelegate:

	case zrpc.RegistService:

	case zrpc.ExecuteDelegate:

	case zrpc.BroadcastDelegate:
	}

	//t.c.SendUnsafeNoCopy(*inboundBuffer.PopData(int(pakSize + 4)))
}

func (e *entity) OnClose() {

}
