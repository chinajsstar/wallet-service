package main

import (
	"net/rpc"
	"api_router/base/service"
	"api_router/base/utils"
	"./handler"
	"fmt"
	"context"
	"time"
	l4g "github.com/alecthomas/log4go"
)

const CobankSrvConfig = "cobank.json"

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	l4g.LoadConfiguration(appDir + "/log.xml")
	defer l4g.Close()

	// create service node
	cfgPath := appDir + "/" + CobankSrvConfig
	fmt.Println("config path:", cfgPath)
	nodeInstance, err:= service.NewServiceNode(cfgPath)
	if nodeInstance == nil || err != nil{
		l4g.Error("Create service node failed: %s", err.Error())
		return
	}
	rpc.Register(nodeInstance)

	// register apis
	cobank := handler.NewCobank()
	if err := cobank.Start(nodeInstance); err != nil{
		l4g.Error("Init service node failed: %s", err.Error())
		return
	}
	service.RegisterNodeApi(nodeInstance, cobank)

	// start ervice node
	ctx, cancel := context.WithCancel(context.Background())
	service.StartNode(ctx, nodeInstance)

	time.Sleep(time.Second*1)
	for ; ;  {
		fmt.Println("Input 'q' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "q" {
			cancel()
			break;
		}
	}

	l4g.Info("Waiting all routine quit...")
	cobank.Stop()
	service.StopNode(nodeInstance)
	l4g.Info("All routine is quit...")
}
