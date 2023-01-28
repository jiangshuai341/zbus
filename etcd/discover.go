package etcd

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
	appName        string
	serviceName    string
	serviceVersion int32
	serviceID      int32
	serviceUrl     []string

	etcdCli               *clientv3.Client //etcd client conn
	etcdEndpoints         []string         //etcd cluster addr
	watchServicesCallback watchServiceCallbackMaps
}

func (d *discover) etcdkey() string {
	return d.appName + "/" + d.serviceName + "/" + strconv.Itoa(int(d.serviceVersion)) + "/" + strconv.Itoa(int(d.serviceID))
}
func (d *discover) etcdval() string {
	var value string
	for index, v := range d.serviceUrl {
		if index == 0 {
			value = v
			continue
		}
		value = value + "," + v
	}
	return value
}
func (d *discover) RegisterAndKeepaliveToETCD(
	appName string,
	serviceName string,
	serviceVersion int32,
	serviceID int32,
	serviceUrl []string,
) {
	if d.etcdCli == nil {
		panic(errors.New("please connect to etcd before regist"))
	}

	d.appName = appName
	d.serviceName = serviceName
	d.serviceVersion = serviceVersion
	d.serviceID = serviceID
	d.serviceUrl = serviceUrl

	leaseResp, err := d.etcdCli.Grant(context.TODO(), 10)

	if err != nil {
		panic("[RegisterAndKeepaliveToETCD] Etcd Call Grant Error:" + err.Error())
	}

	if _, err := d.etcdCli.Put(context.TODO(), d.etcdkey(), d.etcdval(), clientv3.WithLease(leaseResp.ID)); err != nil {
		panic("[RegisterAndKeepaliveToETCD] Etcd Call Put Error:" + err.Error())
	}
	ch, err := d.etcdCli.KeepAlive(context.TODO(), leaseResp.ID)
	if err != nil {
		panic("[RegisterAndKeepaliveToETCD] Etcd Call KeepAlive Error:" + err.Error())
	}

	go func() {
		for resp := range ch {
			if resp != nil {
				log.Infof("Keepalive Client TTL:%d", resp.TTL)
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
func (d *discover) WatchService(serviceName string, watcher IServiceWatcher) {
	keyPrefix := d.appName + "/" + serviceName
	watchchan := d.etcdCli.Watch(context.TODO(), keyPrefix, clientv3.WithPrefix())

	if d.watchServicesCallback.add(serviceName, watcher) {
		go func() {
			for w := range watchchan {
				for _, e := range w.Events {
					name, version, id, ok := parseEtcdKey(e.Kv.Key)
					urls := parseEtcdValue(e.Kv.Value)
					if !ok {
						log.Errorf("[WatchService] parse watch result error key:%s val:%s", string(e.Kv.Key), string(e.Kv.Value))
						continue
					}
					d.watchServicesCallback.execute(name, int32(version), int32(id), urls, e.Type)
				}
			}
		}()
	}
}

func (d *discover) WatchAllService(watcher IServiceWatcher) {
	keyPrefix := d.appName
	watchchan := d.etcdCli.Watch(context.TODO(), keyPrefix, clientv3.WithPrefix())

	if d.watchServicesCallback.add("everyone", watcher) {
		go func() {
			for w := range watchchan {
				for _, e := range w.Events {
					name, version, id, ok := parseEtcdKey(e.Kv.Key)
					urls := parseEtcdValue(e.Kv.Value)
					if !ok {
						log.Errorf("[WatchService] parse watch result error key:%s val:%s", string(e.Kv.Key), string(e.Kv.Value))
						continue
					}
					d.watchServicesCallback.execute(name, int32(version), int32(id), urls, e.Type)
				}
			}
		}()
	}
}

func (d *discover) UnwatchService(serviceName string, watcher IServiceWatcher) {
	d.watchServicesCallback.del(serviceName, watcher)
}

type watchServiceCallbackMaps struct {
	Map    map[string]map[IServiceWatcher]struct{}
	rwLock sync.RWMutex
}
type IServiceWatcher interface {
	OnServiceStatusChange(name string, version int32, serviceId int32, urls []string, eventType mvccpb.Event_EventType)
}
type callbackFun func(name string, version int32, serviceId int32, urls []string, eventType mvccpb.Event_EventType)

func (m *watchServiceCallbackMaps) add(serviceName string, watcher IServiceWatcher) (isFirst bool) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	if m.Map == nil {
		m.Map = make(map[string]map[IServiceWatcher]struct{})
	}
	if m.Map[serviceName] == nil {
		m.Map[serviceName] = make(map[IServiceWatcher]struct{})
		isFirst = true
	}
	m.Map[serviceName][watcher] = struct{}{}
	return
}
func (m *watchServiceCallbackMaps) del(serviceName string, watcher IServiceWatcher) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	if m.Map == nil {
		return
	}
	if m.Map[serviceName] == nil {
		return
	}
	delete(m.Map[serviceName], watcher)
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
	for watcher := range m.Map[serviceName] {
		//todo:操作收到线程池
		go watcher.OnServiceStatusChange(serviceName, serviceVersion, serviceId, urls, eventType)
	}
}
func (m *watchServiceCallbackMaps) reset() {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	m.Map = make(map[string]map[IServiceWatcher]struct{})
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
