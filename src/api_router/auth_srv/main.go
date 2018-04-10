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

const AuthSrvConfig = "auth.json"

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	accountDir := appDir + "/account"
	handler.AuthInstance().Init(accountDir)

	// create service node
	cfgPath := appDir + "/" + AuthSrvConfig
	fmt.Println("config path:", cfgPath)
	nodeInstance, err := service.NewServiceNode(cfgPath)
	if nodeInstance == nil || err != nil{
		fmt.Println("#create service node failed:", err)
		return
	}
	rpc.Register(nodeInstance)

	// register apis
	service.RegisterNodeApi(nodeInstance, handler.AuthInstance())

	// start service node
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