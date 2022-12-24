--[[ testinclude.Vec3

  Automatically generated by the FlatBuffers compiler, do not modify.
  Or modify. I'm a message, not a cop.

  flatc version: 2.0.8

  Declared by  : //include_example.fbs
  Rooting type : testinclude.Vec3 (//include_example.fbs)

--]]

local flatbuffers = require('flatbuffers')

local Vec3 = {}
local mt = {}

function Vec3.New()
  local o = {}
  setmetatable(o, {__index = mt})
  return o
end

function Vec3.GetRootAsVec3(buf, offset)
  if type(buf) == "string" then
    buf = flatbuffers.binaryArray.New(buf)
  end

  local n = flatbuffers.N.UOffsetT:Unpack(buf, offset)
  local o = Vec3.New()
  o:Init(buf, n + offset)
  return o
end

function mt:Init(buf, pos)
  self.view = flatbuffers.view.New(buf, pos)
end

function mt:X()
  local o = self.view:Offset(4)
  if o ~= 0 then
    return self.view:Get(flatbuffers.N.Float32, self.view.pos + o)
  end
  return 0.0
end

function mt:Y()
  local o = self.view:Offset(6)
  if o ~= 0 then
    return self.view:Get(flatbuffers.N.Float32, self.view.pos + o)
  end
  return 0.0
end

function mt:Z()
  local o = self.view:Offset(8)
  if o ~= 0 then
    return self.view:Get(flatbuffers.N.Float32, self.view.pos + o)
  end
  return 0.0
end

function Vec3.Start(builder)
  builder:StartObject(3)
end

function Vec3.AddX(builder, x)
  builder:PrependFloat32Slot(0, x, 0.0)
end

function Vec3.AddY(builder, y)
  builder:PrependFloat32Slot(1, y, 0.0)
end

function Vec3.AddZ(builder, z)
  builder:PrependFloat32Slot(2, z, 0.0)
end

function Vec3.End(builder)
  return builder:EndObject()
end

return Vec3