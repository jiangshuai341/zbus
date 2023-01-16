package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func A() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{
			"localhost:2379",
		},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		// handle error!
	}

	watchchan := cli.Watch(context.Background(), "Test")

	go func() {
		select {
		case i := <-watchchan:
			fmt.Println(i)
		}
	}()
	cli.Put(context.Background(), "Test", "1")
	fmt.Println(cli.Get(context.Background(), "Test"))
	time.Sleep(100 * time.Second)
	_ = cli.Close()
}
