package protorpc

import (
	//"fmt"
	//"github.com/asyoume/protorpc/protobuf"
	//gogo_proto "github.com/gogo/protobuf/proto"
	"log"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	// 注册rpc服务
	//arith := new(Arith)
	//rpc.Register(arith)
	//  监听端口
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	// 开启服务 goroutine
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {

			}
			go ServeConn(conn)
		}
	}()
}
