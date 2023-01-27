package service

import (
	"context"
	"errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strconv"
	"strings"
	"sync"
)

// 服务器集群服务发现模块

type discover struct {
	etcdCli       *clientv3.Client //etcd client conn
	etcdEndpoints []string         //etcd cluster addr

	watchServicesCallback watchServiceCallbackMaps
}

func (s *Service) etcdkey() string {
	return s.appName + "/" + s.serviceName + "/" + strconv.Itoa(int(s.serviceVersion)) + "/" + strconv.Itoa(int(s.serviceID))
}
func (s *Service) etcdval() string {
	var value string
	for index, v := range s.serviceUrl {
		if index == 0 {
			value = v
			continue
		}
		value = value + "," + v
	}
	return value
}
func (s *Service) registerAndKeepaliveToETCD() {
	if s.etcdCli == nil {
		panic(errors.New("please connect to etcd before regist"))
	}

	leaseResp, err := s.etcdCli.Grant(context.TODO(), 10)

	if err != nil {
		panic("[registerAndKeepaliveToETCD] Etcd Call Grant Error:" + err.Error())
	}

	if _, err := s.etcdCli.Put(context.TODO(), s.etcdkey(), s.etcdval(), clientv3.WithLease(leaseResp.ID)); err != nil {
		panic("[registerAndKeepaliveToETCD] Etcd Call Put Error:" + err.Error())
	}
	ch, err := s.etcdCli.KeepAlive(context.TODO(), leaseResp.ID)
	if err != nil {
		panic("[registerAndKeepaliveToETCD] Etcd Call KeepAlive Error:" + err.Error())
	}

	go func() {
		for resp := range ch {
			if resp != nil {
				log.Infof("Keepalive Service TTL:%d", resp.TTL)
			} else {
				log.Errorf("Keepalive Resp Unexpected NULL")
			}
		}
	}()
}
func parseEtcdKey(key []byte) (serviceName string, serviceVersion int, serviceID int, succ bool) {
	strs := strings.Split(string(key), "/")
	if len(strs) < 4 {
		return
	}
	serviceName = strs[1]
	var err error
	serviceVersion, err = strconv.Atoi(strs[2])
	if err != nil {
		return
	}
	serviceID, err = strconv.Atoi(strs[3])
	if err != nil {
		return
	}
	succ = true
	return
}
func parseEtcdValue(value []byte) (urls []string) {
	return strings.Split(string(value), ",")
}
func (s *Service) WatchService(serviceName string, callBack *func(name string, version int32, serviceId int32, urls []string, eventType mvccpb.Event_EventType)) {
	keyPrefix := s.appName + "/" + serviceName
	watchchan := s.etcdCli.Watch(context.TODO(), keyPrefix, clientv3.WithPrefix())

	if s.watchServicesCallback.add(serviceName, (*callbackFun)(callBack)) {
		go func() {
			for w := range watchchan {
				for _, e := range w.Events {
					name, version, id, ok := parseEtcdKey(e.Kv.Key)
					urls := parseEtcdValue(e.Kv.Value)
					if !ok {
						log.Errorf("[WatchService] parse watch result error key:%s val:%s", string(e.Kv.Key), string(e.Kv.Value))
						continue
					}
					s.watchServicesCallback.execute(name, int32(version), int32(id), urls, e.Type)
				}
			}
		}()
	}
}
func (s *Service) UnwatchService(serviceName string, callBack *callbackFun) {
	s.watchServicesCallback.del(serviceName, callBack)
}

type watchServiceCallbackMaps struct {
	Map    map[string]map[*callbackFun]struct{}
	rwLock sync.RWMutex
}

type callbackFun func(name string, version int32, serviceId int32, urls []string, eventType mvccpb.Event_EventType)

func (m *watchServiceCallbackMaps) add(serviceName string, callBack *callbackFun) (isFirst bool) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	if m.Map == nil {
		m.Map = make(map[string]map[*callbackFun]struct{})
	}
	if m.Map[serviceName] == nil {
		m.Map[serviceName] = make(map[*callbackFun]struct{})
		isFirst = true
	}
	m.Map[serviceName][callBack] = struct{}{}
	return
}
func (m *watchServiceCallbackMaps) del(serviceName string, callBack *callbackFun) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	if m.Map == nil {
		return
	}
	if m.Map[serviceName] == nil {
		return
	}
	delete(m.Map[serviceName], callBack)
}
func (m *watchServiceCallbackMaps) execute(serviceName string, serviceVersion int32, serviceId int32, urls []string, eventType mvccpb.Event_EventType) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	if m.Map == nil {
		return
	}
	if m.Map[serviceName] == nil {
		return
	}
	for fun := range m.Map[serviceName] {
		go (*fun)(serviceName, serviceVersion, serviceId, urls, eventType)
	}
}
func (m *watchServiceCallbackMaps) reset() {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	m.Map = make(map[string]map[*callbackFun]struct{})
}

//resp, err := cli.Put(ctx, "", "")
//if err != nil {
//switch err {
//case context.Canceled:
//log.Fatalf("ctx is canceled by another routine: %v", err)
//case context.DeadlineExceeded:
//log.Fatalf("ctx is attached with a deadline is exceeded: %v", err)
//case rpctypes.ErrEmptyKey:
//log.Fatalf("client-side error: %v", err)
//default:
//log.Fatalf("bad cluster endpoints, which are not etcd servers: %v", err)
//}
//}
