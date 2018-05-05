package main

import (
	"api_router/base/service2"
	"api_router/base/utils"
	"fmt"
	"context"
	"time"
	l4g "github.com/alecthomas/log4go"
	"api_router/mytest/rpc2_test/common"
)

const ArithSrvConfig = "testsrv.json"

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	l4g.LoadConfiguration(appDir + "/log.xml")
	defer l4g.Close()

	// create service node
	cfgPath := appDir + "/" + ArithSrvConfig
	fmt.Println("config path:", cfgPath)
	nodeInstance, err:= service2.NewServiceNode(cfgPath)
	if nodeInstance == nil || err != nil{
		l4g.Error("Create service node failed: %s", err.Error())
		return
	}
	//rpc.Register(nodeInstance)

	// register apis
	arith := new(common.Arith)
	service2.RegisterNodeApi(nodeInstance, arith)

	// start ervice node
	ctx, cancel := context.WithCancel(context.Background())
	service2.StartNode(ctx, nodeInstance)

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
	service2.StopNode(nodeInstance)
	l4g.Info("All routine is quit...")
}
