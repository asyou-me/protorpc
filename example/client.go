package main

import (
	"fmt"
	"time"

	"github.com/asyou-me/protorpc"
	"github.com/asyou-me/protorpc/protobuf"
)

func main() {
	cli, err := protorpc.Dial("tcp", "127.0.0.1:1234")

	if err != nil {
		fmt.Println(err)
	}

	defer cli.Close()

	for i := 0; i < 1000; i++ {
		go func(i int64) {
			args := &protobuf.Test{}
			args.A = i
			args.B = i + 1
			reply := new(protobuf.Test)

			err = cli.Call("TestHandler.Test", args, reply)
			if err != nil {
				fmt.Errorf("Add: expected no error but got string %q", err.Error())
			}
			fmt.Println(reply)
			fmt.Println("==============")
		}(int64(i))
	}
	time.Sleep(5 * time.Second)
}
