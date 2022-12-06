#!/bin/bash

#define
#proto 生成输出目录
PATH_OUTPUT=$GOPATH/src
#proto 协议父目录
PATH_PROTO_FATHER_DIR=$GOPATH

SRC_FILE=`find $PATH_PROTO_FATHER_DIR/proto -maxdepth 1 -type f | grep .proto`

#runtime

for FILE in $SRC_FILE
do
    protoc --go_out=plugins=tarsrpc:$PATH_OUTPUT --proto_path=$PATH_PROTO_FATHER_DIR $FILE 
done