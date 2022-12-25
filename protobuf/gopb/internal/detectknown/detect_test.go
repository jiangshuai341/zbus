// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package detectknown_test

import (
	"testing"

	"github.com/jiangshuai341/zbus/protobuf/gopb/internal/detectknown"
	"github.com/jiangshuai341/zbus/protobuf/gopb/reflect/protoreflect"

	fieldmaskpb "github.com/jiangshuai341/zbus/protobuf/gopb/internal/testprotos/fieldmaskpb"
	"github.com/jiangshuai341/zbus/protobuf/gopb/types/descriptorpb"
	"github.com/jiangshuai341/zbus/protobuf/gopb/types/known/anypb"
	"github.com/jiangshuai341/zbus/protobuf/gopb/types/known/durationpb"
	"github.com/jiangshuai341/zbus/protobuf/gopb/types/known/emptypb"
	"github.com/jiangshuai341/zbus/protobuf/gopb/types/known/structpb"
	"github.com/jiangshuai341/zbus/protobuf/gopb/types/known/timestamppb"
	"github.com/jiangshuai341/zbus/protobuf/gopb/types/known/wrapperspb"
	"github.com/jiangshuai341/zbus/protobuf/gopb/types/pluginpb"
)

func TestWhich(t *testing.T) {
	tests := []struct {
		in   protoreflect.FileDescriptor
		want detectknown.ProtoFile
	}{
		{descriptorpb.File_google_protobuf_descriptor_proto, detectknown.Unknown},
		{pluginpb.File_google_protobuf_compiler_plugin_proto, detectknown.Unknown},
		{anypb.File_google_protobuf_any_proto, detectknown.AnyProto},
		{timestamppb.File_google_protobuf_timestamp_proto, detectknown.TimestampProto},
		{durationpb.File_google_protobuf_duration_proto, detectknown.DurationProto},
		{wrapperspb.File_google_protobuf_wrappers_proto, detectknown.WrappersProto},
		{structpb.File_google_protobuf_struct_proto, detectknown.StructProto},
		{fieldmaskpb.File_google_protobuf_field_mask_proto, detectknown.FieldMaskProto},
		{emptypb.File_google_protobuf_empty_proto, detectknown.EmptyProto},
	}

	for _, tt := range tests {
		rangeMessages(tt.in.Messages(), func(md protoreflect.MessageDescriptor) {
			got := detectknown.Which(md.FullName())
			if got != tt.want {
				t.Errorf("Which(%s) = %v, want %v", md.FullName(), got, tt.want)
			}
		})
	}
}

func rangeMessages(mds protoreflect.MessageDescriptors, f func(protoreflect.MessageDescriptor)) {
	for i := 0; i < mds.Len(); i++ {
		md := mds.Get(i)
		if !md.IsMapEntry() {
			f(md)
		}
		rangeMessages(md.Messages(), f)
	}
}
