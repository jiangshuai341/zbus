#!/bin/bash

#define
#proto 生成输出目录
PATH_OUTPUT=/home/jiangshuai/zbus/build/protocal2
#SRC_FILE=`find ./proto -maxdepth 1 -type f | grep .proto`
#
##runtime
#
#for FILE in $SRC_FILE
#do
#    protoc --go_out=plugins=tarsrpc:$PATH_OUTPUT --proto_path=$PATH_PROTO_FATHER_DIR $FILE 
#done


./protoc \
--go_out=$PATH_OUTPUT \
--proto_path=./proto ./proto/test.proto \
--plugin=./protoc-gen-go \
--experimental_allow_proto3_optional

SRC_FILE=`find $PATH_OUTPUT -maxdepth 2 -type f | grep .pb.go`
for FILE in $SRC_FILE 
do
    sed -i "s/github.com\/golang\/protobuf/github.com\/jiangshuai341\/zbus\/protobuf\/gitpb/g"   $FILE
    sed -i "s/google.golang.org\/protobuf/github.com\/jiangshuai341\/zbus\/protobuf\/gopb/g"   $FILE
#    sed -i "s/\"gaia\"/\"common\/autogen\/gaia\"/g" $FILE
#    sed -i "s/\"gaiateam\"/\"common\/autogen\/gaiateam\"/g" $FILE
#    sed -i "s/\"gaiahost\"/\"common\/autogen\/gaiahost\"/g" $FILE
#    sed -i "s/\"pbcommon\"/\"common\/autogen\/pbcommon\"/g" $FILE
done

echo "-- protobuf generated"