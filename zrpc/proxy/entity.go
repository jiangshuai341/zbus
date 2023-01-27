package proxy

import (
	"github.com/jiangshuai341/zbus/zbuffer"
	"github.com/jiangshuai341/zbus/znet/tcp-linux/reactor"
	zrpccommon "github.com/jiangshuai341/zbus/zrpc/common"
)

type entity struct {
	serviceType int32
	entityID    int64
	serviceMap  map[int32]string
	delegateMap map[int32]string
	c           *reactor.Connection
}

func (e *entity) name() {

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

	switch zrpccommon.Cmd(cmd) {
	case zrpccommon.BindDelegate:

	case zrpccommon.RemoteInvoke:

	case zrpccommon.CreateEntity:

	case zrpccommon.DeclareDelegate:

	case zrpccommon.RegistService:

	case zrpccommon.ExecuteDelegate:

	case zrpccommon.BroadcastDelegate:
	}

	//t.c.SendUnsafeNoCopy(*inboundBuffer.PopData(int(pakSize + 4)))
}

func (e *entity) OnClose() {

}
