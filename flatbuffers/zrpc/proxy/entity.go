package proxy

type entity struct {
	serviceType int32
	entityID    int64
	serviceMap  map[int32]string
	delegateMap map[int32]string
}

func (e *entity) name() {

}
