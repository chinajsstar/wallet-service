package main

import (
	"net/rpc"
	"../base/service"
	"../base/utils"
	"./handler"
	"fmt"
	"context"
	"time"
)

const XxxxSrvConfig = "xxxx.json"

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	// create service node
	cfgPath := appDir + "/" + XxxxSrvConfig
	fmt.Println("config path:", cfgPath)
	nodeInstance, err:= service.NewServiceNode(cfgPath)
	if nodeInstance == nil || err != nil{
		fmt.Println("#create service node failed:", err)
		return
	}
	rpc.Register(nodeInstance)

	// register apis
	xxxx := handler.NewXxxx()
	if err := xxxx.Init(); err != nil{
		fmt.Println("#init handler failed:", err)
		return
	}
	service.RegisterNodeApi(nodeInstance, xxxx)

	// start ervice node
	ctx, cancel := context.WithCancel(context.Background())
	service.StartNode(ctx, nodeInstance)

	time.Sleep(time.Second*1)
	for ; ;  {
		fmt.Println("Input 'quit' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "quit" {
			cancel()
			break;
		}
	}

	fmt.Println("Waiting all routine quit...")
	service.StopNode(nodeInstance)
	fmt.Println("All routine is quit...")
}
