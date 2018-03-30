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
	"strings"
	"errors"
)

const ArithSrvName = "arith"
const ArithSrvVersion = "v1"
const (
	GateWayAddr = "127.0.0.1:8081"
	SrvAddr = "127.0.0.1:8090"
)
var g_apisMap = make(map[string]service.CallNodeApi)

// 注册方法
func callArithFunction(req *data.SrvDispatchData, ack *data.SrvDispatchAckData) error{
	var err error
	h := g_apisMap[strings.ToLower(req.SrvArgv.Function)]
	if h != nil {
		err = h(req, ack)
	}else{
		err = errors.New("not find api")
	}

	fmt.Println("callNodeApi req: ", *req)
	fmt.Println("callNodeApi ack: ", *ack)

	return err
}

func main() {
	wg := &sync.WaitGroup{}

	// 创建节点
	nodeInstance, _:= service.NewServiceNode(ArithSrvName, ArithSrvVersion)
	nodeInstance.RegisterData.Addr = SrvAddr
	arith := new(handler.Arith)
	arith.RegisterApi(&nodeInstance.RegisterData.Functions, &g_apisMap)
	nodeInstance.Handler = callArithFunction

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
