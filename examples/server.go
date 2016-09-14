package main

import (
	"fmt"
	"net"

	"github.com/asyou-me/protorpc"
	"github.com/asyou-me/protorpc/types"
)

// 用于关闭测试服务的管道
var closeChan chan struct{}

// 开启一个测试用的服务
func server(addr string) {
	closeChan = make(chan struct{})

	// 注册rpc服务
	h := new(TestHandler)
	server := protorpc.NewServer()

	server.Register(h)
	server.Auth(func(p *protorpc.AuthorizationHeader) error {
		return nil
	})

	//  监听端口
	l, e := net.Listen("tcp", addr)
	if e != nil {
		fmt.Println("listen error:", e)
	}

	// 开启服务 goroutine
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("conn:", err)
			} else {
				fmt.Println("开启一个新的连接:", conn.RemoteAddr())
			}
			go func() {
				server.ServeConn(conn)
			}()
		}
	}()
}

// TestHandler 测试服务
type TestHandler struct {
}

// Test 测试服务 方法
func (h *TestHandler) Test(arg *types.Test, reply *types.Test) error {
	reply.A = arg.A
	reply.B = arg.B
	reply.C = reply.A + reply.B
	return nil
}

func main() {
	server("127.0.0.1:30015")
	<-closeChan
}
