package main

import (
	"github.com/asyoume/protorpc/protobuf"
	//gogo_proto "github.com/gogo/protobuf/proto"
	"fmt"
	"github.com/asyoume/protorpc"
	"net"
	"net/rpc"
)

func main() {

	var close_chan chan struct{} = make(chan struct{})

	// 注册rpc服务
	h := new(TestHandler)
	rpc.Register(h)

	//  监听端口
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		fmt.Println("listen error:", e)
	}

	// 开启服务 goroutine
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("conn:", err)
			}
			go protorpc.ServeConn(conn, protobuf.HanderMap_)
		}
	}()
	<-close_chan
}

// 测试服务
type TestHandler struct {
}

func (h *TestHandler) Test(arg *protobuf.Test, reply *protobuf.Test) error {
	fmt.Println("test", arg)
	reply.Id = "jieguo"
	return nil
}
