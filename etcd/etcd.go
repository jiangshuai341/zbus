package etcd

import (
	"context"
	"github.com/jiangshuai341/zbus/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var log = logger.GetLogger("AA")

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

	watchchan := cli.Watch(context.Background(), "/Test/", clientv3.WithPrefix())

	go func() {
		for {
			select {
			case i := <-watchchan:
				log.Infof("%+v", i)
			}
		}
	}()
	clientv3.NewKV(cli)
	cli.Put(context.Background(), "/Test/sub/", "1")
	cli.Put(context.Background(), "/Test/", "2")
	time.Sleep(100 * time.Second)
	_ = cli.Close()
}

//func RegisterServiceToETCD(ServiceTarget string, value string) {
//	dir := strings.TrimRight(ServiceTarget, "/") + "/"
//
//	client, err := clientv3.New(clientv3.Config{
//		Endpoints:   []string{"localhost:2379"},
//		DialTimeout: 5 * time.Second,
//	})
//	if err != nil {
//		panic(err)
//	}
//
//	kv := clientv3.NewKV(client)
//	lease := clientv3.NewLease(client)
//	var curLeaseId clientv3.LeaseID = 0
//
//	for {
//		if curLeaseId == 0 {
//			leaseResp, err := lease.Grant(context.TODO(), 10)
//			if err != nil {
//				panic(err)
//			}
//
//			key := ServiceTarget + fmt.Sprintf("%d", leaseResp.ID)
//			if _, err := kv.Put(context.TODO(), key, value, clientv3.WithLease(leaseResp.ID)); err != nil {
//				panic(err)
//			}
//			curLeaseId = leaseResp.ID
//		} else {
//			// 续约租约，如果租约已经过期将curLeaseId复位到0重新走创建租约的逻辑
//			if _, err := lease.KeepAliveOnce(context.TODO(), curLeaseId); err == rpctypes.ErrLeaseNotFound {
//				curLeaseId = 0
//				continue
//			}
//		}
//		time.Sleep(time.Duration(1) * time.Second)
//	}
//}
