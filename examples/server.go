package main

import (
	"fmt"
	"net"

	"github.com/asyou-me/protorpc/types"

	"github.com/asyou-me/protorpc"
)

func main() {

	var closeChan = make(chan struct{})

	// 注册rpc服务
	h := new(TestHandler)
	server := protorpc.NewServer()

	server.Register(h)
	server.Auth(func(p *protorpc.AuthorizationHeader) error {
		return nil
	})

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
			go func() {
				server.ServeConn(conn)
			}()
		}
	}()
	<-closeChan
}

// TestHandler 测试服务
type TestHandler struct {
}

var i = 0

// Test 测试服务 方法
func (h *TestHandler) Test(arg *types.Test, reply *types.Test) error {
	//fmt.Println("test", arg)
	reply.A = arg.A
	reply.B = arg.B
	i++
	reply.C = reply.A + reply.B
	fmt.Println("i:", i)
	return nil
}
