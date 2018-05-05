package main

import (
	//"api_router/base/service"
	service "api_router/base/service2"
	"fmt"
	"time"
	"context"
	"api_router/base/utils"
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

	// start service center
	ctx, cancel := context.WithCancel(context.Background())
	service.StartCenter(ctx, centerInstance)

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
	service.StopCenter(centerInstance)
	l4g.Info("All routine is quit...")
}

