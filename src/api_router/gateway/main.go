package main

import (
	"net/rpc"
	"../base/service"
	"fmt"
	"time"
	"context"
)

const ServiceCenterName = "center"

func main() {
	// 创建服务中心
	centerInstance,_ := service.NewServiceCenter(ServiceCenterName)
	centerInstance.HttpPort = ":8080"
	centerInstance.TcpPort = ":8081"

	// 注册RPC接口
	centerInstance.RpcServer = rpc.NewServer()
	centerInstance.RpcServer.Register(centerInstance)

	// 启动服务中心
	ctx, cancel := context.WithCancel(context.Background())
	go service.StartCenter(ctx, centerInstance)

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
	centerInstance.Wg.Wait()
	fmt.Println("All routine is quit...")
}

