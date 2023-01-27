package zrpccommon

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
