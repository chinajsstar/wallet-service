package main

import (
	//"api_router/base/service"
	service "api_router/base/service2"
	"api_router/base/data"
	"api_router/arith_srv/handler"
	"fmt"
	"context"
	"time"
	"strconv"
	l4g "github.com/alecthomas/log4go"
	"api_router/base/config"
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
	cfgDir := config.GetBastionPayConfigDir()

	l4g.LoadConfiguration(cfgDir + "/log.xml")
	defer l4g.Close()

	// create service node
	cfgPath := cfgDir + "/" + ArithSrvConfig
	fmt.Println("config path:", cfgPath)
	nodeInstance, err:= service.NewServiceNode(cfgPath)
	if nodeInstance == nil || err != nil{
		l4g.Error("Create service node failed: %s", err.Error())
		return
	}

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
