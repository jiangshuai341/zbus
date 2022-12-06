package hash

import (
	"easyframe/logger"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

const DEFAULT_REPLICAS = 32

var log = logger.GetLogger("hash")

type SortKeys []uint32
type HashRing struct {
	ServerName string
	Nodes      map[uint32]string
	Keys       SortKeys
}

func Init(nodes []string, Hashring *HashRing) {
	Hashring.Nodes = make(map[uint32]string)
	Hashring.Keys = SortKeys{}
	if nodes == nil {
		return
	}
	for _, node := range nodes { //实际物理节点
		for i := 0; i < DEFAULT_REPLICAS; i++ { //虚拟节点的映射
			str := node + strconv.Itoa(i)
			Hashring.Nodes[HashStr(str)] = node
			Hashring.Keys = append(Hashring.Keys, HashStr(str))
			//log.GetLogger("TestHash").Debug("hr.Keys:", hr.Keys, "Virtual->Real", hr.Nodes)
		}
	}
	log.Infof("HashInit::Physical Node:%+v virtual node num:%d", nodes, len(Hashring.Keys))
	sort.Sort(Hashring.Keys)
}
func (hr *HashRing) ReplacePhysicalNode(nodes []string) { //并非添加而是更新，内部是更新和添加处理
	if nodes == nil {
		return
	}
	for _, node := range nodes { //实际物理节点
		for i := 0; i < DEFAULT_REPLICAS; i++ { //虚拟节点的映射
			str := node + strconv.Itoa(i)
			hashval := HashStr(str)
			_, ok := hr.Nodes[hashval]
			if ok && hr.Nodes[hashval] == node {
				continue
			}
			if ok && hr.Nodes[hashval] != node {
				log.Errorf("Name:%s Physical->virtual hash error：old:%s", hr.ServerName, hr.Nodes[hashval], node)
				continue
			}
			hr.Nodes[hashval] = node
			hr.Keys = append(hr.Keys, hashval)
			log.Infof("[Server: %s]New virtual Node HashValue:%d Node:%s", hr.ServerName, hashval, node)
		}
	}
	sort.Sort(hr.Keys)
}
func (hr *HashRing) DelPhysicalNode(nodes []string) {
	if nodes == nil {
		return
	}
	for _, node := range nodes {
		for i := 0; i < DEFAULT_REPLICAS; i++ {
			str := node + strconv.Itoa(i)
			delete(hr.Nodes, HashStr(str))
			index, ok := hr.findVirtualNodeInKeys(HashStr(str))
			if !ok {
				continue
			}
			log.Infof("[DelPhysicalNode] Server:%s Node:%s", hr.ServerName, node)
			hr.Keys = append(hr.Keys[:index], hr.Keys[index+1:]...)
		}
	}

	sort.Sort(hr.Keys)
}

func (hr *HashRing) GetNode(key string) string {
	if len(hr.Nodes) == 0 {
		return ""
	}
	hash := HashStr(key)
	i := hr.get_position(hash)
	return hr.Nodes[hr.Keys[i]]
}

var keyValCache sync.Map

func (hr *HashRing) GetNodeInt64(key int64) string {
	if len(hr.Nodes) == 0 {
		return ""
	}
	val, ok := keyValCache.Load(key)
	var hashval uint32 = 0
	if ok {
		hashval = val.(uint32)
	} else {
		hashval = HashStr(strconv.FormatInt(key, 10))
		keyValCache.Store(key, hashval)
	}
	i := hr.get_position(hashval)
	return hr.Nodes[hr.Keys[i]]
}

// //==============================工具函数====================================
func (hr *HashRing) findVirtualNodeInKeys(hashvalue uint32) (int, bool) {
	for index, value := range hr.Keys {
		if value == hashvalue {
			return index, true
		}
	}
	return 0, false
}

func HashStr(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

func (hr *HashRing) get_position(hashvalue uint32) int {
	i := sort.Search(len(hr.Keys), func(i int) bool { //二分法搜索找到[0, n)区间内最小的满足f(i)==true的值i
		return hr.Keys[i] >= hashvalue // 数组 a =｛0 ，1 ，2， 5， 10 ，15｝  { a[i]>=4 return i=3 } {a[i]>=14 return 5}
	})

	if i < len(hr.Keys) {
		if i == len(hr.Keys)-1 {
			return 0
		} else {
			return i
		}
	} else {
		return len(hr.Keys) - 1
	}
}
func (sk SortKeys) Len() int {
	return len(sk)
}

func (sk SortKeys) Less(i, j int) bool {
	return sk[i] < sk[j]
}

func (sk SortKeys) Swap(i, j int) {
	sk[i], sk[j] = sk[j], sk[i]
}
