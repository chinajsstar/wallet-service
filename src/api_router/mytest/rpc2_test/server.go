package main

import (
	"github.com/cenkalti/rpc2"
	"fmt"
	"net"
	"api_router/mytest/rpc2_test/common"
	"time"
)

type ClientInfo struct{
	Client *rpc2.Client
}

var clients map[string]*ClientInfo

func main()  {
	clients = make(map[string]*ClientInfo)

	srv := rpc2.NewServer()

	srv.OnConnect(func(client *rpc2.Client) {
		fmt.Println("connect:")
	})

	srv.OnDisconnect(func(client *rpc2.Client) {
		fmt.Println("disconnect:")
		for k, v := range clients  {
			if v.Client == client {
				delete(clients, k)
				fmt.Println("disconnect:", k)
				break
			}
		}
	})

	srv.Handle("register", func(client *rpc2.Client, args *common.Register, res *string) error {
		// client regist

		clients[args.FuncName] = &ClientInfo{Client:client}
		fmt.Println("add a client, funcname = ", args.FuncName)

		*res = "ok"
		return nil
	})
	srv.Handle("unregister", func(client *rpc2.Client, args *common.Register, res *string) error {
		// client regist

		delete(clients, args.FuncName)
		fmt.Println("remove a client, funcname = ", args.FuncName)

		*res = "ok"
		return nil
	})

	srv.Handle("dispatch", func(client *rpc2.Client, args *common.Args, reply *common.Reply) error {
		// Reversed call (server to client)

		if c, ok := clients[args.FuncName]; ok && c != nil {
			c.Client.Call("call", args, reply)

			fmt.Println("dispath to call:", args, reply)
		}else{
			fmt.Println("no clt, dispath to call:", args, reply)
		}
		return nil
	})

	go func() {
		lis, _ := net.Listen("tcp", "127.0.0.1:5000")
		srv.Accept(lis)
	}()

	time.Sleep(time.Second*1)
	for ; ;  {
		fmt.Println("Input 'q' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "q" {
			break;
		} else if input == "test"{
			fmt.Println("test call:", input)
			for _, v := range clients  {

				for i := 0; i < 10000; i++{
					var reply common.Reply
					v.Client.Call("call", &common.Args{FuncName:input}, &reply)

					fmt.Println("direct call:", reply)
				}
			}
			fmt.Println("end direct call:", input)
		}else {
			fmt.Println("direct call:", input)
			if c, ok := clients[input]; ok && c != nil {
				var reply common.Reply
				c.Client.Call("call", &common.Args{FuncName:input}, &reply)

				fmt.Println("direct call:", reply)
			}
			fmt.Println("end direct call:", input)
		}
	}
}
