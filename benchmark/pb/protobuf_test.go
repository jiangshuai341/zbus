package pb

import (
	"github.com/jiangshuai341/zbus/benchmark/pb/test"
	"github.com/jiangshuai341/zbus/protobuf/gopb/proto"
	"testing"
)

var m = test.Monster{
	Mana:      1000,
	Hp:        1000,
	Name:      "hhhhhhhhhhh",
	Friendly:  true,
	Inventory: make([]byte, 128),
	Color:     test.Color_Blue,
	Weapons:   make([]*test.Weapon, 128),
	Path:      make([]*test.Vec3, 128),
}

// goos: linux
// goarch: amd64
// pkg: github.com/jiangshuai341/zbus/build/pb
// cpu: AMD Ryzen 9 5900X 12-Core Processor
// BenchmarkMarshal
// BenchmarkMarshal-24       372901              2892 ns/op
// PASS
func BenchmarkMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(&m)
	}
}

// goos: linux
// goarch: amd64
// pkg: github.com/jiangshuai341/zbus/build/pb
// cpu: AMD Ryzen 9 5900X 12-Core Processor
// BenchmarkMarshalAndUnmarshal
// BenchmarkMarshalAndUnmarshal-24            54027             21862 ns/op
// PASS
func BenchmarkMarshalAndUnmarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data, _ := proto.Marshal(&m)
		temp := &test.Monster{}
		_ = proto.Unmarshal(data, temp)
	}
}
