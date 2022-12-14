// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proto

import (
	"github.com/jiangshuai341/zbus/protobuf/gopb/encoding/protowire"
	"github.com/jiangshuai341/zbus/protobuf/gopb/internal/encoding/messageset"
	"github.com/jiangshuai341/zbus/protobuf/gopb/internal/errors"
	"github.com/jiangshuai341/zbus/protobuf/gopb/internal/flags"
	"github.com/jiangshuai341/zbus/protobuf/gopb/reflect/protoreflect"
	"github.com/jiangshuai341/zbus/protobuf/gopb/reflect/protoregistry"
)

func sizeMessageSet(m protoreflect.Message) (size int) {
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		size += messageset.SizeField(fd.Number())
		size += protowire.SizeTag(messageset.FieldMessage)
		size += protowire.SizeBytes(sizeMessage(v.Message()))
		return true
	})
	size += messageset.SizeUnknown(m.GetUnknown())
	return size
}

func marshalMessageSet(b []byte, m protoreflect.Message, o MarshalOptions) ([]byte, error) {
	if !flags.ProtoLegacy {
		return b, errors.New("no support for message_set_wire_format")
	}
	var err error
	o.rangeFields(m, func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		b, err = marshalMessageSetField(b, fd, v, o)
		return err == nil
	})
	if err != nil {
		return b, err
	}
	return messageset.AppendUnknown(b, m.GetUnknown())
}

func marshalMessageSetField(b []byte, fd protoreflect.FieldDescriptor, value protoreflect.Value, o MarshalOptions) ([]byte, error) {
	b = messageset.AppendFieldStart(b, fd.Number())
	b = protowire.AppendTag(b, messageset.FieldMessage, protowire.BytesType)
	b = protowire.AppendVarint(b, uint64(o.Size(value.Message().Interface())))
	b, err := o.marshalMessage(b, value.Message())
	if err != nil {
		return b, err
	}
	b = messageset.AppendFieldEnd(b)
	return b, nil
}

func unmarshalMessageSet(b []byte, m protoreflect.Message, o UnmarshalOptions) error {
	if !flags.ProtoLegacy {
		return errors.New("no support for message_set_wire_format")
	}
	return messageset.Unmarshal(b, false, func(num protowire.Number, v []byte) error {
		err := unmarshalMessageSetField(m, num, v, o)
		if err == errUnknown {
			unknown := m.GetUnknown()
			unknown = protowire.AppendTag(unknown, num, protowire.BytesType)
			unknown = protowire.AppendBytes(unknown, v)
			m.SetUnknown(unknown)
			return nil
		}
		return err
	})
}

func unmarshalMessageSetField(m protoreflect.Message, num protowire.Number, v []byte, o UnmarshalOptions) error {
	md := m.Descriptor()
	if !md.ExtensionRanges().Has(num) {
		return errUnknown
	}
	xt, err := o.Resolver.FindExtensionByNumber(md.FullName(), num)
	if err == protoregistry.NotFound {
		return errUnknown
	}
	if err != nil {
		return errors.New("%v: unable to resolve extension %v: %v", md.FullName(), num, err)
	}
	xd := xt.TypeDescriptor()
	if err := o.unmarshalMessage(v, m.Mutable(xd).Message()); err != nil {
		return err
	}
	return nil
}
