// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dynamicpb_test

import (
	"testing"

	"github.com/jiangshuai341/zbus/protobuf/gopb/proto"
	pref "github.com/jiangshuai341/zbus/protobuf/gopb/reflect/protoreflect"
	preg "github.com/jiangshuai341/zbus/protobuf/gopb/reflect/protoregistry"
	"github.com/jiangshuai341/zbus/protobuf/gopb/testing/prototest"
	"github.com/jiangshuai341/zbus/protobuf/gopb/types/dynamicpb"

	testpb "github.com/jiangshuai341/zbus/protobuf/gopb/internal/testprotos/test"
	test3pb "github.com/jiangshuai341/zbus/protobuf/gopb/internal/testprotos/test3"
)

func TestConformance(t *testing.T) {
	for _, message := range []proto.Message{
		(*testpb.TestAllTypes)(nil),
		(*test3pb.TestAllTypes)(nil),
		(*testpb.TestAllExtensions)(nil),
	} {
		mt := dynamicpb.NewMessageType(message.ProtoReflect().Descriptor())
		prototest.Message{}.Test(t, mt)
	}
}

func TestDynamicExtensions(t *testing.T) {
	for _, message := range []proto.Message{
		(*testpb.TestAllExtensions)(nil),
	} {
		mt := dynamicpb.NewMessageType(message.ProtoReflect().Descriptor())
		prototest.Message{
			Resolver: extResolver{},
		}.Test(t, mt)
	}
}

type extResolver struct{}

func (extResolver) FindExtensionByName(field pref.FullName) (pref.ExtensionType, error) {
	xt, err := preg.GlobalTypes.FindExtensionByName(field)
	if err != nil {
		return nil, err
	}
	return dynamicpb.NewExtensionType(xt.TypeDescriptor().Descriptor()), nil
}

func (extResolver) FindExtensionByNumber(message pref.FullName, field pref.FieldNumber) (pref.ExtensionType, error) {
	xt, err := preg.GlobalTypes.FindExtensionByNumber(message, field)
	if err != nil {
		return nil, err
	}
	return dynamicpb.NewExtensionType(xt.TypeDescriptor().Descriptor()), nil
}

func (extResolver) RangeExtensionsByMessage(message pref.FullName, f func(pref.ExtensionType) bool) {
	preg.GlobalTypes.RangeExtensionsByMessage(message, func(xt pref.ExtensionType) bool {
		return f(dynamicpb.NewExtensionType(xt.TypeDescriptor().Descriptor()))
	})
}
