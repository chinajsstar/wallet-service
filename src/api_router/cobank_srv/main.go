package main

import (
	//"api_router/base/service"
	service "api_router/base/service2"
	"api_router/base/utils"
	"api_router/cobank_srv/handler"
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

	onLineDir := ""//"/Users/henly.liu/wallet-service/src/bastionpay_tools/web_online/data"
	// register apis
	cobank := handler.NewCobank(onLineDir)
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
