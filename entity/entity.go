package entity

type Type uint8

const (
	Service Type = iota
	Player
	MatchTeam
	MatchRoom
	DsAgent
	DsServer
)

type Entity struct {
	entityType Type
}

func CreateEntity(t Type, key int64, rpcImp any) *Entity {

	return nil
}

func (e *Entity) RPC() {

}
