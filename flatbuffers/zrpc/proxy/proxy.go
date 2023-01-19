package proxy

import (
	"github.com/jiangshuai341/zbus/flatbuffers/zrpc/common"
	"github.com/jiangshuai341/zbus/toolkit"
	"github.com/jiangshuai341/zbus/zbuffer"
	"github.com/jiangshuai341/zbus/znet/tcp-linux/reactor"
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

var ioReactor *reactor.Reactor
var accepter *reactor.Accepter

func init() {
	port, err := toolkit.GetFreePort()
	if err != nil {
		panic("rpc proxy service toolkit.GetFreePort failed Err:" + err.Error())
	}

	listeners, err := reactor.ActiveListener("0.0.0.0:"+strconv.Itoa(port), 1, OnAccept)
	if err != nil {
		panic("rpc proxy service reactor.ActiveListener failed Err:" + err.Error())
	}
	accepter = listeners[0]

	err = addListenUDS()
	if err != nil {
		panic("rpc proxy service addListenUDS failed Err:" + err.Error())
	}

	ioReactor, err = reactor.NewReactor()
	if err != nil {
		panic("rpc proxy service new reactor failed Err:" + err.Error())
	}
}

func OnAccept(conn *reactor.Connection) {
	conn.INetHandle = &NetTask{c: conn}
	err := ioReactor.AddConn(conn)
	if err != nil {
		return
	}
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

	cmd, err := inboundBuffer.PeekInt(4, 4)
	if err != nil {
		return
	}

	switch common.Cmd(cmd) {
	case common.BindDelegate:

	case common.RemoteInvoke:

	case common.CreateEntity:

	case common.DeclareDelegate:

	case common.RegistService:

	case common.ExecuteDelegate:

	case common.BroadcastDelegate:
	}

	//t.c.SendUnsafeNoCopy(*inboundBuffer.PopData(int(pakSize + 4)))
}

func (t *NetTask) OnClose() {

}
