// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package Example

import (
	"github.com/jiangshuai341/zbus/benchmark/fb/testinclude"
	flatbuffers "github.com/jiangshuai341/zbus/flatbuffers"
)

type MonsterT struct {
	Pos       *testinclude.Vec3T   `json:"pos"`
	Mana      int32                `json:"mana"`
	Hp        int32                `json:"hp"`
	Name      string               `json:"name"`
	Inventory []byte               `json:"inventory"`
	Color     Color                `json:"color"`
	Weapons   []*WeaponT           `json:"weapons"`
	Path      []*testinclude.Vec3T `json:"path"`
}

func (t *MonsterT) Pack(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	if t == nil {
		return 0
	}
	posOffset := t.Pos.Pack(builder)
	nameOffset := builder.CreateString(t.Name)
	inventoryOffset := flatbuffers.UOffsetT(0)
	if t.Inventory != nil {
		inventoryOffset = builder.CreateByteString(t.Inventory)
	}
	weaponsOffset := flatbuffers.UOffsetT(0)
	if t.Weapons != nil {
		weaponsLength := len(t.Weapons)
		weaponsOffsets := make([]flatbuffers.UOffsetT, weaponsLength)
		for j := 0; j < weaponsLength; j++ {
			weaponsOffsets[j] = t.Weapons[j].Pack(builder)
		}
		MonsterStartWeaponsVector(builder, weaponsLength)
		for j := weaponsLength - 1; j >= 0; j-- {
			builder.PrependUOffsetT(weaponsOffsets[j])
		}
		weaponsOffset = builder.EndVector(weaponsLength)
	}
	pathOffset := flatbuffers.UOffsetT(0)
	if t.Path != nil {
		pathLength := len(t.Path)
		pathOffsets := make([]flatbuffers.UOffsetT, pathLength)
		for j := 0; j < pathLength; j++ {
			pathOffsets[j] = t.Path[j].Pack(builder)
		}
		MonsterStartPathVector(builder, pathLength)
		for j := pathLength - 1; j >= 0; j-- {
			builder.PrependUOffsetT(pathOffsets[j])
		}
		pathOffset = builder.EndVector(pathLength)
	}
	MonsterStart(builder)
	MonsterAddPos(builder, posOffset)
	MonsterAddMana(builder, t.Mana)
	MonsterAddHp(builder, t.Hp)
	MonsterAddName(builder, nameOffset)
	MonsterAddInventory(builder, inventoryOffset)
	MonsterAddColor(builder, t.Color)
	MonsterAddWeapons(builder, weaponsOffset)
	MonsterAddPath(builder, pathOffset)
	return MonsterEnd(builder)
}

func (rcv *Monster) UnPackTo(t *MonsterT) {
	t.Pos = rcv.Pos(nil).UnPack()
	t.Mana = rcv.Mana()
	t.Hp = rcv.Hp()
	t.Name = string(rcv.Name())
	t.Inventory = rcv.InventoryBytes()
	t.Color = rcv.Color()
	weaponsLength := rcv.WeaponsLength()
	t.Weapons = make([]*WeaponT, weaponsLength)
	for j := 0; j < weaponsLength; j++ {
		x := Weapon{}
		rcv.Weapons(&x, j)
		t.Weapons[j] = x.UnPack()
	}
	pathLength := rcv.PathLength()
	t.Path = make([]*testinclude.Vec3T, pathLength)
	for j := 0; j < pathLength; j++ {
		x := testinclude.Vec3{}
		rcv.Path(&x, j)
		t.Path[j] = x.UnPack()
	}
}

func (rcv *Monster) UnPack() *MonsterT {
	if rcv == nil {
		return nil
	}
	t := &MonsterT{}
	rcv.UnPackTo(t)
	return t
}

type Monster struct {
	_tab flatbuffers.Table
}

func GetRootAsMonster(buf []byte, offset flatbuffers.UOffsetT) *Monster {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Monster{}
	x.Init(buf, n+offset)
	return x
}

func GetSizePrefixedRootAsMonster(buf []byte, offset flatbuffers.UOffsetT) *Monster {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &Monster{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func (rcv *Monster) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Monster) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Monster) Pos(obj *testinclude.Vec3) *testinclude.Vec3 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(testinclude.Vec3)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *Monster) Mana() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 150
}

func (rcv *Monster) MutateMana(n int32) bool {
	return rcv._tab.MutateInt32Slot(6, n)
}

func (rcv *Monster) Hp() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 100
}

func (rcv *Monster) MutateHp(n int32) bool {
	return rcv._tab.MutateInt32Slot(8, n)
}

func (rcv *Monster) Name() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Monster) Inventory(j int) byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.GetByte(a + flatbuffers.UOffsetT(j*1))
	}
	return 0
}

func (rcv *Monster) InventoryLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *Monster) InventoryBytes() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Monster) MutateInventory(j int, n byte) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.MutateByte(a+flatbuffers.UOffsetT(j*1), n)
	}
	return false
}

func (rcv *Monster) Color() Color {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		return Color(rcv._tab.GetInt8(o + rcv._tab.Pos))
	}
	return 2
}

func (rcv *Monster) MutateColor(n Color) bool {
	return rcv._tab.MutateInt8Slot(16, int8(n))
}

func (rcv *Monster) Weapons(obj *Weapon, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *Monster) WeaponsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *Monster) Path(obj *testinclude.Vec3, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(20))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *Monster) PathLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(20))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func MonsterStart(builder *flatbuffers.Builder) {
	builder.StartObject(9)
}
func MonsterAddPos(builder *flatbuffers.Builder, pos flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(pos), 0)
}
func MonsterAddMana(builder *flatbuffers.Builder, mana int32) {
	builder.PrependInt32Slot(1, mana, 150)
}
func MonsterAddHp(builder *flatbuffers.Builder, hp int32) {
	builder.PrependInt32Slot(2, hp, 100)
}
func MonsterAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(name), 0)
}
func MonsterAddInventory(builder *flatbuffers.Builder, inventory flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(5, flatbuffers.UOffsetT(inventory), 0)
}
func MonsterStartInventoryVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(1, numElems, 1)
}
func MonsterAddColor(builder *flatbuffers.Builder, color Color) {
	builder.PrependInt8Slot(6, int8(color), 2)
}
func MonsterAddWeapons(builder *flatbuffers.Builder, weapons flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(7, flatbuffers.UOffsetT(weapons), 0)
}
func MonsterStartWeaponsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func MonsterAddPath(builder *flatbuffers.Builder, path flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(8, flatbuffers.UOffsetT(path), 0)
}
func MonsterStartPathVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func MonsterEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
