package main

import (
	//"api_router/base/service"
	service "api_router/base/service2"
	"fmt"
	"time"
	"context"
	l4g "github.com/alecthomas/log4go"
	"api_router/base/config"
	"bastionpay_api/utils"
)

const ServiceGatewayConfig = "gateway.json"

func main() {
	cfgDir := config.GetBastionPayConfigDir()

	l4g.LoadConfiguration(cfgDir + "/log.xml")
	defer l4g.Close()

	defer utils.PanicPrint()

	cfgPath := cfgDir + "/" + ServiceGatewayConfig
	fmt.Println("config path:", cfgPath)

	// create service center
	gatewayInstance, err := service.NewServiceGateway(cfgPath)
	if gatewayInstance == nil || err != nil {
		l4g.Error("Create service center failed: %s", err.Error())
		return
	}

	// start service center
	ctx, cancel := context.WithCancel(context.Background())
	service.StartCenter(ctx, gatewayInstance)

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
	service.StopCenter(gatewayInstance)
	l4g.Info("All routine is quit...")
}

