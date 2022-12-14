// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/jiangshuai341/zbus/protobuf/gopb/encoding/prototext"
	"github.com/jiangshuai341/zbus/protobuf/gopb/internal/fieldnum"
	"github.com/jiangshuai341/zbus/protobuf/gopb/proto"

	"github.com/jiangshuai341/zbus/protobuf/gopb/types/descriptorpb"
)

func TestAnnotations(t *testing.T) {
	sourceFile, err := ioutil.ReadFile("testdata/annotations/annotations.pb.go")
	if err != nil {
		t.Fatal(err)
	}
	metaFile, err := ioutil.ReadFile("testdata/annotations/annotations.pb.go.meta")
	if err != nil {
		t.Fatal(err)
	}
	gotInfo := &descriptorpb.GeneratedCodeInfo{}
	if err := prototext.Unmarshal(metaFile, gotInfo); err != nil {
		t.Fatalf("can't parse meta file: %v", err)
	}

	wantInfo := &descriptorpb.GeneratedCodeInfo{}
	for _, want := range []struct {
		prefix, text, suffix string
		path                 []int32
	}{{
		"type ", "AnnotationsTestEnum", " int32",
		[]int32{fieldnum.FileDescriptorProto_EnumType, 0},
	}, {
		"\t", "AnnotationsTestEnum_ANNOTATIONS_TEST_ENUM_VALUE", " AnnotationsTestEnum = 0",
		[]int32{fieldnum.FileDescriptorProto_EnumType, 0, fieldnum.EnumDescriptorProto_Value, 0},
	}, {
		"type ", "AnnotationsTestMessage", " struct {",
		[]int32{fieldnum.FileDescriptorProto_MessageType, 0},
	}, {
		"\t", "AnnotationsTestField", " ",
		[]int32{fieldnum.FileDescriptorProto_MessageType, 0, fieldnum.DescriptorProto_Field, 0},
	}, {
		"func (x *AnnotationsTestMessage) ", "GetAnnotationsTestField", "() string {",
		[]int32{fieldnum.FileDescriptorProto_MessageType, 0, fieldnum.DescriptorProto_Field, 0},
	}} {
		s := want.prefix + want.text + want.suffix
		pos := bytes.Index(sourceFile, []byte(s))
		if pos < 0 {
			t.Errorf("source file does not contain: %v", s)
			continue
		}
		begin := pos + len(want.prefix)
		end := begin + len(want.text)
		wantInfo.Annotation = append(wantInfo.Annotation, &descriptorpb.GeneratedCodeInfo_Annotation{
			Path:       want.path,
			Begin:      proto.Int32(int32(begin)),
			End:        proto.Int32(int32(end)),
			SourceFile: proto.String("cmd/protoc-gen-go/testdata/annotations/annotations.proto"),
		})
	}
	if !proto.Equal(gotInfo, wantInfo) {
		t.Errorf("unexpected annotations for annotations.proto; got:\n%v\nwant:\n%v", gotInfo, wantInfo)
	}
}
