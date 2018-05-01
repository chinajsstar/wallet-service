package main

import (
	"net/rpc"
	"api_router/base/service"
	"api_router/base/utils"
	"api_router/base/data"
	"api_router/arith_srv/handler"
	"fmt"
	"context"
	"time"
	"strconv"
	l4g "github.com/alecthomas/log4go"
)

const ArithSrvConfig = "arith.json"

func testPush(node *service.ServiceNode)  {
	for i := 0; i < 50; i++ {
		time.Sleep(time.Second*5)

		pData := data.UserRequestData{}
		pData.Method.Version = "v1"
		pData.Method.Srv = "push"
		pData.Method.Function = "pushdata"
		pData.Argv.UserKey = "3b7ecf3b-c605-4c4f-ac2b-2155d4186cd8"
		pData.Argv.Message = "abcd=" + strconv.Itoa(i)

		res := data.UserResponseData{}
		node.InnerCallByEncrypt(&pData, &res)

		fmt.Println(res)
	}
}

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	l4g.LoadConfiguration(appDir + "/log.xml")
	defer l4g.Close()

	// create service node
	cfgPath := appDir + "/" + ArithSrvConfig
	fmt.Println("config path:", cfgPath)
	nodeInstance, err:= service.NewServiceNode(cfgPath)
	if nodeInstance == nil || err != nil{
		l4g.Error("Create service node failed: %s", err.Error())
		return
	}
	rpc.Register(nodeInstance)

	// register apis
	arith := new(handler.Arith)
	service.RegisterNodeApi(nodeInstance, arith)

	// start ervice node
	ctx, cancel := context.WithCancel(context.Background())
	service.StartNode(ctx, nodeInstance)

	// start test push
	go testPush(nodeInstance)

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
	service.StopNode(nodeInstance)
	l4g.Info("All routine is quit...")
}
