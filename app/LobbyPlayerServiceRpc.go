package lobby

import (
	"github.com/jiangshuai341/zbus/build/fb/lobby"
	vector "github.com/jiangshuai341/zbus/build/fb/vector"
)

type ILobbyPlayerNetDiver interface {
	SendByIdName()
}
type ILobbyPlayerRpcRunner interface {
	SubmitTask(fun func())
}

type LobbyPlayerEntityProxy struct {
	serviceName int32
	hashKey     int64
	netdriver   ILobbyPlayerNetDiver
}

func NewLobbyPlayerEntityProxy(hashKey int64, netdriver ILobbyPlayerNetDiver) *LobbyPlayerEntityProxy {
	return &LobbyPlayerEntityProxy{
		serviceName: 123132,
		hashKey:     hashKey,
		netdriver:   netdriver,
	}
}
func (p *LobbyPlayerEntityProxy) StoreSync(in lobby.Monster) (*vector.Vec3, error) {
	return nil, nil
}
func (p *LobbyPlayerEntityProxy) StoreAsync(in lobby.Monster, callback func(*vector.Vec3, error)) {

}

func (p *LobbyPlayerEntityProxy) BindMaxHpChange(in lobby.Monster, callback func()) {

}

type ILobbyPlayerServiceImp interface {
	Store()
}

type LobbyPlayerEntity struct {
	serviceName int32
	hashKey     int64
	imp         ILobbyPlayerServiceImp
	netdriver   ILobbyPlayerNetDiver
	runner      ILobbyPlayerRpcRunner
}

func NewLobbyPlayerEntity(hashKey int64, imp ILobbyPlayerServiceImp, netdriver ILobbyPlayerNetDiver, runner ILobbyPlayerRpcRunner) *LobbyPlayerEntity {
	//在Gproxy上注册
	//在RpcProcessor上注册
	return &LobbyPlayerEntity{
		serviceName: 123132,
		hashKey:     hashKey,
		imp:         imp,
		runner:      runner,
		netdriver:   netdriver,
	}
}

func (e *LobbyPlayerEntity) Dispatch(fun int32, req []byte, resp *[]byte) {
	switch fun {
	case 1121313:
		e.imp.Store()
	}
	return
}

func (e *LobbyPlayerEntity) Close() {

}

func (e *LobbyPlayerEntity) MaxHpChangeBroadcast(in lobby.Monster, callback func(*vector.Vec3, error)) {

}
