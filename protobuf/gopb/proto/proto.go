// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proto

import (
	"github.com/jiangshuai341/zbus/protobuf/gopb/internal/errors"
	"github.com/jiangshuai341/zbus/protobuf/gopb/reflect/protoreflect"
)

// Message is the top-level interface that all messages must implement.
// It provides access to a reflective view of a message.
// Any implementation of this interface may be used with all functions in the
// protobuf module that accept a Message, except where otherwise specified.
//
// This is the v2 interface definition for protobuf messages.
// The v1 interface definition is "github.com/jiangshuai341/zbus/protobuf/gitpb/proto".Message.
//
// To convert a v1 message to a v2 message,
// use "github.com/jiangshuai341/zbus/protobuf/gitpb/proto".MessageV2.
// To convert a v2 message to a v1 message,
// use "github.com/jiangshuai341/zbus/protobuf/gitpb/proto".MessageV1.
type Message = protoreflect.ProtoMessage

// Error matches all errors produced by packages in the protobuf module.
//
// That is, errors.Is(err, Error) reports whether an error is produced
// by this module.
var Error error

func init() {
	Error = errors.Error
}
