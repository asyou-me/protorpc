# protobuf RPC for go


## EXAMPLES
server:
```
package main

import (
	"fmt"
	"net"

	"github.com/asyou-me/protorpc"
	"github.com/asyou-me/protorpc/types"
)

func main() {
	// 注册rpc服务
	h := new(TestHandler)
	server := protorpc.NewServer()

	server.Register(h)
	server.Auth(func(p *protorpc.AuthorizationHeader) error {
		return nil
	})

	//  监听端口
	l, e := net.Listen("tcp", "127.0.0.1:30015")
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
```

client:
```
package main

import (
	"fmt"
	"net"

	"github.com/asyou-me/protorpc"
	"github.com/asyou-me/protorpc/types"
)

func main() {
	cli, err := protorpc.Dial("tcp", "127.0.0.1:30015")
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
}
```

client_pool:
```
package main

import (
	"fmt"
	"net"

	"github.com/asyou-me/protorpc"
	"github.com/asyou-me/protorpc/types"
)

func main() {
	closeChan = make(chan struct{}, 100)
	var addr = "127.0.0.1:30015"
	cli := Pool{
		Dial: func() (*protorpc.Client, error) {
			cli, err := protorpc.Dial("tcp", addr)
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

					closeChan <- struct{}{}
				}
			}(int64(i))
			time.Sleep(10 * time.Millisecond)
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
```

## TODO

## INSTALL
To install it, you must first have Go (at least version 1.3.3) installed (see http://golang.org/doc/install). Go 1.4.2, 1.5.4, 1.6.3 and 1.7 are continuously tested.

Next, install the standard protocol buffer implementation from https://github.com/google/protobuf. Most versions from 2.3.1 should not give any problems, but 2.5.0, 2.6.1 and 3 are continuously tested.

Next, install [gogo/protobuf](https://github.com/gogo/protobuf)

Final, install asyou-me/protorpc
```
go get github.com/asyou-me/protorpc
```

## CHANGE LOG

## Benchmark Test

