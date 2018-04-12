package main

import (
	"../base/service"
	"fmt"
	"time"
	"context"
	"../base/utils"
	"net/rpc"
	l4g "github.com/alecthomas/log4go"
)

const ServiceGatewayConfig = "gateway.json"

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	l4g.LoadConfiguration(appDir + "/log.xml")
	defer l4g.Close()

	cfgPath := appDir + "/" + ServiceGatewayConfig
	fmt.Println("config path:", cfgPath)

	// create service center
	centerInstance, err := service.NewServiceCenter(cfgPath)
	if centerInstance == nil || err != nil {
		l4g.Error("Create service center failed: %s", err.Error())
		return
	}
	rpc.Register(centerInstance)

	// start service center
	ctx, cancel := context.WithCancel(context.Background())
	service.StartCenter(ctx, centerInstance)

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

	l4g.Info("Waiting all routine quit...")
	service.StopCenter(centerInstance)
	l4g.Info("All routine is quit...")
}

