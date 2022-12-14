// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package detectknown provides functionality for detecting well-known types
// and identifying them by name.
package detectknown

import "github.com/jiangshuai341/zbus/protobuf/gopb/reflect/protoreflect"

type ProtoFile int

const (
	Unknown ProtoFile = iota
	AnyProto
	TimestampProto
	DurationProto
	WrappersProto
	StructProto
	FieldMaskProto
	EmptyProto
)

var wellKnownTypes = map[protoreflect.FullName]ProtoFile{
	"google.protobuf.Any":         AnyProto,
	"google.protobuf.Timestamp":   TimestampProto,
	"google.protobuf.Duration":    DurationProto,
	"google.protobuf.BoolValue":   WrappersProto,
	"google.protobuf.Int32Value":  WrappersProto,
	"google.protobuf.Int64Value":  WrappersProto,
	"google.protobuf.UInt32Value": WrappersProto,
	"google.protobuf.UInt64Value": WrappersProto,
	"google.protobuf.FloatValue":  WrappersProto,
	"google.protobuf.DoubleValue": WrappersProto,
	"google.protobuf.BytesValue":  WrappersProto,
	"google.protobuf.StringValue": WrappersProto,
	"google.protobuf.Struct":      StructProto,
	"google.protobuf.ListValue":   StructProto,
	"google.protobuf.Value":       StructProto,
	"google.protobuf.FieldMask":   FieldMaskProto,
	"google.protobuf.Empty":       EmptyProto,
}

// Which identifies the proto file that a well-known type belongs to.
func Which(s protoreflect.FullName) ProtoFile {
	return wellKnownTypes[s]
}
