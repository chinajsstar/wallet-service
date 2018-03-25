package main

import (
	"net/rpc"
	"../base/service"
	"../data"
	"./handler"
	"fmt"
	"context"
	"time"
	"sync"
)

const AuthSrvName = "auth"
const AuthSrvVersion = "v1"
const (
	GateWayAddr = "127.0.0.1:8081"
	SrvAddr = "127.0.0.1:8091"
)

// 注册方法
func callAuthFunction(req *data.ServiceCenterDispatchData, ack *data.ServiceCenterDispatchAckData){
	// TODO:
	ack.Err = 0
	ack.Value = req.Argv

	fmt.Println("callNodeApi req: ", *req)
	fmt.Println("callNodeApi ack: ", *ack)
}

func main() {
	wg := &sync.WaitGroup{}

	// 创建节点
	nodeInstance, _:= service.NewServiceNode(AuthSrvName, AuthSrvVersion)
	nodeInstance.RegisterData.Addr = SrvAddr
	nodeInstance.RegisterData.RegisterFunction(new(handler.Auth))
	nodeInstance.Handler = callAuthFunction

	nodeInstance.ServiceCenterAddr = GateWayAddr
	rpc.Register(nodeInstance)

	// 启动节点服务
	ctx, cancel := context.WithCancel(context.Background())
	nodeInstance.Start(ctx, wg)

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
