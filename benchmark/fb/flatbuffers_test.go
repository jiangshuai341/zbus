package fb

import (
	Example2 "github.com/jiangshuai341/zbus/benchmark/fb/Example"
	"github.com/jiangshuai341/zbus/benchmark/fb/testinclude"
	"github.com/jiangshuai341/zbus/flatbuffers"
	"testing"
)

var m = Example2.MonsterT{
	Mana:      1000,
	Hp:        1000,
	Name:      "hhhhhhhhhhh",
	Inventory: make([]byte, 128),
	Color:     Example2.ColorBlue,
	Weapons:   make([]*Example2.WeaponT, 128),
	Path:      make([]*testinclude.Vec3T, 128),
}

// goos: linux
// goarch: amd64
// pkg: github.com/jiangshuai341/zbus/build/fb
// cpu: AMD Ryzen 9 5900X 12-Core Processor
// BenchmarkMarshal
// BenchmarkMarshal-24       488415              2321 ns/op
// PASS

func BenchmarkMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m.Pack(flatbuffers.NewBuilder(1024))
	}
}

// goos: linux
// goarch: amd64
// pkg: github.com/jiangshuai341/zbus/build/fb
// cpu: AMD Ryzen 9 5900X 12-Core Processor
// BenchmarkMarshalAndUnmarshal
// BenchmarkMarshalAndUnmarshal-24           466560              2391 ns/op
// PASS
func BenchmarkMarshalAndUnmarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bl := flatbuffers.NewBuilder(1024)
		_ = m.Pack(bl)
		temp := Example2.GetRootAsMonster(bl.Bytes, 0)
		temp.UnPack()
	}
}
