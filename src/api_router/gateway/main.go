package main

import (
	"../base/service"
	"fmt"
	"time"
	"context"
	"sync"
	"../base/config"
	"../base/utils"
)

const ServiceCenterConfig = "center.json"

func main() {
	var err error
	cc := config.ConfigCenter{}
	if err = cc.Load(utils.GetRunDir()+"/config/"+ServiceCenterConfig); err != nil{
		err = cc.Load(utils.GetCurrentDir() + "/config/" + ServiceCenterConfig)
	}
	if err != nil {
		return
	}
	fmt.Println("config:", cc)

	wg := &sync.WaitGroup{}

	// 创建服务中心
	centerInstance,_ := service.NewServiceCenter(cc.CenterName, ":"+cc.Port, ":"+cc.WsPort, ":"+cc.CenterPort)

	// 启动服务中心
	ctx, cancel := context.WithCancel(context.Background())
	centerInstance.Start(ctx, wg)

	time.Sleep(time.Second*2)
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
	wg.Wait()
	fmt.Println("All routine is quit...")
}

