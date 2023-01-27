package example

import (
	"fmt"
	"github.com/jiangshuai341/zbus/service"
	"github.com/jiangshuai341/zbus/znet/tcp-linux/reactor"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"testing"
	"time"
)

func testOnAccepet(conn *reactor.Connection) {
	return
}
func TestDiscover(t *testing.T) {
	gateway := service.NewService([]string{"localhost:2379"},
		"gaia", "gateway",
		1, 2,
		[]string{"tcp://127.0.0.1:8888"},
		testOnAccepet)

	callback := func(name string, version int32, serviceId int32, urls []string, eventType mvccpb.Event_EventType) {
		fmt.Println(name, version, serviceId, urls, eventType)
	}
	gateway.WatchService("lobby", &callback)

	service.NewService([]string{"localhost:2379"},
		"gaia", "lobby",
		1, 3,
		[]string{"tcp://127.0.0.1:8889"},
		testOnAccepet)

	service.NewService([]string{"localhost:2379"},
		"gaia", "lobby",
		1, 4,
		[]string{"tcp://127.0.0.1:8890"},
		testOnAccepet)

	time.Sleep(10000 * time.Second)
}
