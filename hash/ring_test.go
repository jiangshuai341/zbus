package hash

import (
	"strconv"
	"testing"
)

type service struct {
	member int32
}

func (s *service) GetID() []byte {
	return []byte(strconv.Itoa(int(s.member)))
}

func addService(ring *HashRing[*service], services ...int32) {
	for _, v := range services {
		var serviceA *service = &service{
			member: v,
		}
		ring.Add(serviceA)
	}

}

func BenchmarkHashRing_Get(b *testing.B) {
	ring := NewHashRing[*service]("test", 32)
	addService(ring, 112312, 2123141, 423423, 12312, 234234, 235235)
	var a = []byte{123, 123, 124, 145, 2}
	for i := 0; i < b.N; i++ {
		ring.Get(a)
	}
}

func FuzzHashRing_Get(f *testing.F) {
	ring := NewHashRing[*service]("test", 32)
	addService(ring, 112312, 2123141, 423423, 12312, 234234, 235235)

	f.Add("3213432")
	f.Fuzz(func(t *testing.T, hashKey string) {
		ring.Get([]byte(hashKey))
	})
}
