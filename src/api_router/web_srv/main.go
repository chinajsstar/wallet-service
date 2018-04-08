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

const WebSrvConfig = "web.json"

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	// create service node
	cfgPath := appDir + "/" + WebSrvConfig
	fmt.Println("config path:", cfgPath)
	nodeInstance, err:= service.NewServiceNode(cfgPath)
	if nodeInstance == nil || err != nil{
		fmt.Println("#create service node failed:", err)
		return
	}
	rpc.Register(nodeInstance)

	// register apis
	web := new(handler.Web)
	if err := web.Init(nodeInstance); err != nil{
		fmt.Println("#init service node failed:", err)
		return
	}
	service.RegisterNodeApi(nodeInstance, web)

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
