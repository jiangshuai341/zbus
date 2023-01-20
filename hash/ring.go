package hash

import (
	"hash/crc32"
	"sort"
)

type INode interface {
	GetID() []byte
}
type node[T INode] struct {
	sortKey uint32
	hashKey uint32
	data    T
}
type HashRing[T INode] struct {
	name                 string
	nodes                []node[T]
	numberOfVirtualNodes int
}

func NewHashRing[T INode](name string, numberOfVirtualNodes int) *HashRing[T] {
	return &HashRing[T]{
		name:                 name,
		numberOfVirtualNodes: numberOfVirtualNodes,
		nodes:                make([]node[T], 0),
	}
}

func (h *HashRing[T]) Add(n T) {
	temp := node[T]{
		data:    n,
		hashKey: crc32.ChecksumIEEE(n.GetID()),
	}
	for i := 0; i < h.numberOfVirtualNodes; i++ { //虚拟节点的映射
		temp.sortKey = crc32.ChecksumIEEE(append(n.GetID(), byte(i)))
		h.nodes = append(h.nodes, temp)
	}
	sort.Sort(nodeSlice[T](h.nodes))
}

func (h *HashRing[T]) Del(n T) {
	hashKey := crc32.ChecksumIEEE(n.GetID())
	for i, hasBeenDeleted := 0, 0; i < len(h.nodes) && hasBeenDeleted < h.numberOfVirtualNodes; {
		if h.nodes[i].hashKey == hashKey {
			hasBeenDeleted++
			h.nodes = append(h.nodes[0:i], h.nodes[i+1:]...)
		} else {
			i++
		}
	}
}
func (h *HashRing[T]) Get(key []byte) (ret T) {
	if len(h.nodes) == 0 {
		return
	}
	hashvalue := crc32.ChecksumIEEE(key)
	i := sort.Search(len(h.nodes), func(i int) bool { //二分法搜索找到[0, n)区间内最小的满足f(i)>=true的值i
		return h.nodes[i].sortKey >= hashvalue // 数组 a =｛0 ，1 ，2， 5， 10 ，15｝  { a[i]>=4 return i=3 } {a[i]>=14 return 5}
	})

	if i == len(h.nodes) {
		i = 0
	}
	return h.nodes[i].data
}

type nodeSlice[T INode] []node[T]

func (n nodeSlice[T]) Len() int {
	return len(n)
}
func (n nodeSlice[T]) Less(i, j int) bool {
	return n[i].sortKey < n[j].sortKey
}
func (n nodeSlice[T]) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}
