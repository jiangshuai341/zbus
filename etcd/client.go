package etcd

import (
	"context"
	"github.com/jiangshuai341/zbus/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var log = logger.GetLogger("")

type baseInfo struct {
	appName        string
	serviceName    string
	serviceVersion int32
	serviceID      int32
	serviceUrl     []string
}

type Client struct {
	discover
}

func NewClient(etcdEndpoints []string) *Client {
	cfg := clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 5 * time.Second,
	}

	etcdConn, err := clientv3.New(cfg)
	if err != nil {
		panic("[NewService] connect to etcd failed err:" + err.Error())
	}

	return &Client{
		discover: discover{
			etcdCli: etcdConn,
		},
	}
}

func (s *Client) Close() {
	_, _ = s.etcdCli.Delete(context.TODO(), s.etcdkey())
	_ = s.etcdCli.Close()

	s.watchServicesCallback.reset()
}
