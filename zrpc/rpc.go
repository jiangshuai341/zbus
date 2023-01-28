package zrpc

import "github.com/jiangshuai341/zbus/zpool/coroutinepool"

type Cmd int32

const (
	BindDelegate Cmd = iota
	RemoteInvoke

	CreateEntity
	DeclareDelegate
	RegistService
	ExecuteDelegate
	BroadcastDelegate
)

type NetDriver interface {
}

func (r *RpcProcessor) RemoteInvokeSync(hashKey int64, fun int32, req []byte, resp *[]byte, driver NetDriver) {

}
func (r *RpcProcessor) RemoteInvokeAsync(hashKey int64, fun int32, req []byte, resp *[]byte, driver NetDriver, callback func(bytes []byte, err error)) {

}
func (r *RpcProcessor) ParseMsg(hashKey int64, fun int32, req []byte, resp *[]byte, driver NetDriver) {

}

type RpcProcessor struct {
	routinePool coroutinepool.RoutinePool
	driver      NetDriver
}

func testA() {

	//routinePool := coroutinepool.New()
}
