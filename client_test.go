package protorpc

import (
	"fmt"
	"testing"

	"github.com/asyou-me/protorpc/types"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	fmt.Println("开始测试 Client")
	var addr = "127.0.0.1:30015"
	server(addr)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	cli, err := Dial("tcp", addr, "")
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	args := &types.Test{}
	args.A = 1
	args.B = 1
	reply := new(types.Test)
	err = cli.Call("TestHandler.Test", args, reply)
	if err != nil {
		t.Fatal("Error:", err.Error())
	}
	assert.Equal(t, args.A, reply.A, "args.A 不等于 reply.A ,多线程数据紊乱")
	assert.Equal(t, args.B, reply.B, "args.B 不等于 reply.B ,多线程数据紊乱")
	assert.Equal(t, reply.C, int64(2), "reply.C 不等于2 ,服务端返回结果错误")
	closeChan <- struct{}{}
}
