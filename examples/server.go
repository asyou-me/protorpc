package main

import (
	"fmt"

	"net"
	"net/rpc"

	"github.com/asyou-me/protorpc/protobuf"

	"github.com/asyou-me/protorpc"
)

func main() {

	var close_chan chan struct{} = make(chan struct{})

	// 注册rpc服务
	h := new(TestHandler)
	rpc.Register(h)

	//  监听端口
	l, e := net.Listen("tcp", ":1236")
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
			go protorpc.ServeConn(conn)
		}
	}()
	<-close_chan
}

// 测试服务
type TestHandler struct {
}

var i = 0

func (h *TestHandler) Test(arg *protobuf.Test, reply *protobuf.Test) error {
	//fmt.Println("test", arg)
	reply.A = arg.A
	reply.B = arg.B
	i++
	reply.C = reply.A + reply.B
	fmt.Println("i:", i)
	return nil
}
