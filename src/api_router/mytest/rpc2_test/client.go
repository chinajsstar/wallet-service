package main

import (
	"github.com/cenkalti/rpc2"
	"fmt"
	"net"
	"api_router/mytest/rpc2_test/common"
	"time"
)

func main()  {
	fmt.Println("Input funcname to register...")
	var funcName string
	fmt.Scanln(&funcName)

	conn, _ := net.Dial("tcp", "127.0.0.1:5000")
	clt := rpc2.NewClient(conn)

	clt.Handle("call", func(client *rpc2.Client, args *common.Args, reply *common.Reply) error {
		fmt.Println("funcname= ", args.FuncName)
		if args.FuncName == funcName {
			*reply = "ok"
		}else{
			*reply = "no"
		}
		return nil
	})
	go func() {
		clt.Run()
		<- clt.DisconnectNotify()
		fmt.Println("disconnect")
	}()
	defer clt.Close()

	var rep common.Reply
	clt.Call("register", common.Register{funcName}, &rep)
	fmt.Println("register result:", rep)

	time.Sleep(time.Second*1)
	for ; ;  {
		fmt.Println("Input 'q' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "q" {
			break;
		}else if input == "test"{
			fmt.Println("test call:", input)

			for i := 0; i < 10000; i++{
				var reply common.Reply
				clt.Call("dispatch", &common.Args{FuncName:input}, &reply)

				fmt.Println("direct call:", reply)
			}
			fmt.Println("end direct call:", input)
		}else{
			err := clt.Call("dispatch", common.Args{input}, &rep)
			fmt.Println("dispatch result:", rep, err)
		}
	}

	clt.Call("unregister", common.Register{funcName}, &rep)
	fmt.Println("unregister result:", rep)
}
