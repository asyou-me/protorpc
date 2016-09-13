package protorpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/asyou-me/protorpc/types"
	"github.com/stretchr/testify/assert"
)

// 用于关闭测试服务的管道
var PoolChan chan struct{}

func TestClientPool(t *testing.T) {
	fmt.Println("开始测试 ClientPool")
	closeChan = make(chan struct{}, 100)
	var addr = "127.0.0.1:30016"
	server(addr)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	cli := Pool{
		Dial: func() (*Client, error) {
			cli, err := Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			return cli, nil
		}, Max: 100,
	}
	cli.Init()

	go func() {
		for i := 0; i < 10; i++ {
			go func(i int64) {
				for j := 0; j < 100; j++ {
					args := &types.Test{}
					args.A = i*100 + int64(j)
					args.B = 1
					reply := new(types.Test)
					err := cli.Call("TestHandler.Test", args, reply)
					if err != nil {
						fmt.Println("Error:", err.Error())
						t.Fatal("Error:", err.Error())
					}
					assert.Equal(t, args.A, reply.A, "args.A 不等于 reply.A ,多线程数据紊乱")
					assert.Equal(t, args.B, reply.B, "args.B 不等于 reply.B ,多线程数据紊乱")
					assert.Equal(t, reply.C, args.A+args.B, "reply.C 结果错误 ,服务端返回结果错误")

					closeChan <- struct{}{}
				}
			}(int64(i))
			time.Sleep(20 * time.Millisecond)
		}
	}()

	var j = 0
	for {
		select {
		case _ = <-closeChan:
			j = j + 1
		}

		if j == 999 {
			break
		}
	}
}
