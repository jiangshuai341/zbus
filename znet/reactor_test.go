package znet

import (
	"github.com/jiangshuai341/zbus/znet/reactor"
	"io"
	"net"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func TestListen(t *testing.T) {
	runServer(1)
	time.Sleep(1000000 * time.Second)
}

func TestClient(t *testing.T) {
	runClient(1024*10*10, 1)
	time.Sleep(1000000 * time.Second)
}

func runServer(num int) {
	_, _ = reactor.ActiveListener("0.0.0.0:9999", num, OnAccept)
}

func runClient(dataBlockSize int, clientNum int) {
	for i := 0; i < clientNum; i++ {
		conn, _ := net.Dial("tcp", "0.0.0.0:9999")
		var syncCtx sync.WaitGroup
		syncCtx.Add(1)

		var pakData = func(data []byte) []byte {
			var ret = make([]byte, 4, 4+len(data))
			*(*int32)(unsafe.Pointer(&ret[0])) = int32(len(data))
			ret = append(ret, data...)
			return ret
		}

		go func() {
			var tempRead = make([]byte, dataBlockSize+4)
			var tempWrite = make([]byte, dataBlockSize)
			for {
				_, err := conn.Write(pakData(tempWrite))
				if err != nil {
					break
				}
				_, err = io.ReadAtLeast(conn, tempRead, len(tempRead))
				if err != nil {
					break
				}
			}
		}()
	}
}
