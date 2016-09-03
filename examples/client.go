package main

import (
	"fmt"
	"github.com/asyoume/protorpc"
	"github.com/asyoume/protorpc/protobuf"
	"net/rpc"
)

func main() {
	//client()
	client_pool()
}

func client() {
	cli, err := protorpc.Dial("tcp", "127.0.0.1:1236")
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
			//fmt.Println(reply)
			//fmt.Println("==============")
		}(int64(i))
	}
}

func client_pool() {
	var close_chan chan struct{} = make(chan struct{})
	cli := protorpc.Pool{
		Dial: func() (*rpc.Client, error) {
			cli, err := protorpc.Dial("tcp", "127.0.0.1:1236")
			if err != nil {
				return nil, err
			}
			return cli, nil
		}, Max: 100,
	}
	cli.Init()

	for i := 0; i < 1000000; i++ {
		go func(i int64) {
			args := &protobuf.Test{}
			args.A = i
			args.B = i + 1
			reply := new(protobuf.Test)

			err := cli.Call("TestHandler.Test", args, reply)
			if err != nil {
				fmt.Errorf("Add: expected no error but got string %q", err.Error())
			}
			//fmt.Println(reply)
			//fmt.Println("==============")
			if i%10000 == 0 {
				fmt.Println(i)
			}
		}(int64(i))
	}
	<-close_chan
}
