package main

import (
	"../base/service"
	"fmt"
	"time"
	"context"
	"../base/utils"
	"net/rpc"
)

const ServiceGatewayConfig = "gateway.json"

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	cfgPath := appDir + "/" + ServiceGatewayConfig
	fmt.Println("config path:", cfgPath)

	// create service center
	centerInstance, err := service.NewServiceCenter(cfgPath)
	if centerInstance == nil || err != nil {
		fmt.Println("#create service center failed:", err)
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

	fmt.Println("Waiting all routine quit...")
	service.StopCenter(centerInstance)
	fmt.Println("All routine is quit...")
}

