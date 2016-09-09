#!/bin/bash

# 源于网络，用于获取当前shell文件的路径
SOURCE="$0"
while [ -h "$SOURCE"  ]; do 
    DIR="$( cd -P "$( dirname "$SOURCE"  )" && pwd  )"
    SOURCE="$(readlink "$SOURCE")"
    [[ $SOURCE != /*  ]] && SOURCE="$DIR/$SOURCE" 
done
DIR="$( cd -P "$( dirname "$SOURCE"  )" && pwd  )"

gogoproto_path="$GOPATH/src/"

protoc3 "$DIR/api.proto" --gofast_out="$DIR/" --proto_path=${gogoproto_path} --proto_path="$DIR"