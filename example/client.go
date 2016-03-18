package main

import (
	"fmt"
	"github.com/asyoume/protorpc"
	"github.com/asyoume/protorpc/protobuf"
)

func main() {
	cli, err := protorpc.Dial("tcp", "127.0.0.1:1234", protobuf.HanderMap)

	if err != nil {
		fmt.Println(err)
	}

	defer cli.Close()

	for i := 0; i < 100; i++ {

		args := &protobuf.Test{}
		args.Id = "qinqiu"
		reply := new(protobuf.Test)

		err = cli.Call("TestHandler.Test", args, reply)
		if err != nil {
			fmt.Errorf("Add: expected no error but got string %q", err.Error())
		}
		fmt.Println(reply)
		fmt.Println("==============")
	}
}

type Reply struct {
	C int
}
