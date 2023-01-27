package service

import (
	"context"
	"github.com/jiangshuai341/zbus/logger"
	"github.com/jiangshuai341/zbus/znet/tcp-linux/reactor"
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

type Service struct {
	discover
	baseInfo
	network
}

func NewService(
	etcdEndpoints []string,
	appName string,
	serviceName string,
	serviceVersion int32,
	serviceID int32,
	serviceUrl []string,
	onAccept func(conn *reactor.Connection),
) (s *Service) {
	if etcdEndpoints == nil || len(etcdEndpoints) == 0 ||
		serviceUrl == nil || len(serviceUrl) == 0 ||
		serviceID == 0 {
		panic("[NewService] check [etcdEndpoints] [serviceUrl] [serviceID]")
	}

	cfg := clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 5 * time.Second,
	}

	etcdConn, err := clientv3.New(cfg)
	if err != nil {
		panic("[NewService] connect to etcd failed err:" + err.Error())
	}

	s = &Service{}
	s.etcdCli = etcdConn
	s.etcdEndpoints = etcdEndpoints
	s.appName = appName
	s.serviceName = serviceName
	s.serviceVersion = serviceVersion
	s.serviceID = serviceID
	s.serviceUrl = serviceUrl
	s.onAccept = onAccept

	//json.Unmarshal()
	s.registerAndKeepaliveToETCD()
	s.networkStart()
	return
}

func (s *Service) Close() {
	_, _ = s.etcdCli.Delete(context.TODO(), s.etcdkey())
	_ = s.etcdCli.Close()

	s.watchServicesCallback.reset()
}
