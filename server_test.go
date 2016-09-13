package protorpc

import (
	"fmt"
	"net"

	"github.com/asyou-me/protorpc/types"
)

// 用于关闭测试服务的管道
var closeChan chan struct{}

// 开启一个测试用的服务
func server(addr string) {
	closeChan = make(chan struct{})

	// 注册rpc服务
	h := new(TestHandler)
	server := NewServer()

	server.Register(h)
	server.Auth(func(p *AuthorizationHeader) error {
		return nil
	})

	//  监听端口
	l, e := net.Listen("tcp", addr)
	if e != nil {
		fmt.Println("listen error:", e)
	}

	// 开启服务 goroutine
	var j = 0
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("conn:", err)
			}
			go func() {
				server.ServeConn(conn)
			}()
			select {
			case _ = <-closeChan:
				j++
			}
			if j > 0 {
				l.Close()
				break
			}
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
