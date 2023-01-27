package proxy

import (
	"github.com/jiangshuai341/zbus/service"
	"github.com/jiangshuai341/zbus/toolkit"
	"github.com/jiangshuai341/zbus/znet/tcp-linux/reactor"
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

var netDriver *reactor.Reactor
var accepter *reactor.Accepter

func init() {
	var err error
	if accepter, err = reactor.NewListener(OnAccept); err != nil {
		panic("rpc proxy service reactor.NewListener failed Err:" + err.Error())
	}
	port, err := toolkit.GetFreePort()
	if err != nil {
		panic("rpc proxy service toolkit.GetFreePort failed Err:" + err.Error())
	}
	if err = accepter.ListenUrl("tcp://" + "0.0.0.0:" + strconv.Itoa(port)); err != nil {
		panic("rpc proxy service accepter.ListenUrl failed Err:" + err.Error())
	}
	if runtime.GOOS == "linux" {
		if err = accepter.ListenUrl("unix://" + "/tmp/hhhhh.sock"); err != nil {
			panic("rpc proxy service listen unix socket failed Err:" + err.Error())
		}
	}

	netDriver, err = reactor.NewReactor()
	if err != nil {
		panic("rpc proxy service new reactor failed Err:" + err.Error())
	}

	service.NewService([]string{"localhost:2379"},
		"zbus", "gproxy",
		1, 2,
		[]string{"tcp://127.0.0.1:8888"})

}

func OnAccept(conn *reactor.Connection) {
	conn.INetHandle = &entity{c: conn}
	err := netDriver.AddConn(conn)
	if err != nil {
		return
	}
}
