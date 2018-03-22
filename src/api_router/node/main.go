package main

import (
	"net/rpc"
	"../base/service"
	"../base/common"
	"./function"
	"fmt"
	"context"
	"time"
)

const ServiceNodeName = "node"
const ServiceNodeVersion = "v1"

func callNodeApi(req *common.ServiceCenterDispatchData, ack *common.ServiceCenterDispatchAckData){

	fmt.Println("callNodeApi req: ", *req)
	fmt.Println("callNodeApi ack: ", *ack)
}

func main() {
	// 创建节点
	nodeInstance, _:= service.NewServiceNode(ServiceNodeName, ServiceNodeVersion)
	nodeInstance.RegisterData.Addr = "127.0.0.1:8090"
	nodeInstance.RegisterData.RegisterApi(new(function.MyFunc1))
	nodeInstance.RegisterData.RegisterApi(new(function.MyFunc2))

	nodeInstance.ServiceCenterAddr = "127.0.0.1:8081"
	nodeInstance.Handler = callNodeApi

	rpc.Register(nodeInstance)

	// 启动节点服务
	ctx, cancel := context.WithCancel(context.Background())
	go service.StartNode(ctx, nodeInstance)

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
	nodeInstance.Wg.Wait()
	fmt.Println("All routine is quit...")
}
