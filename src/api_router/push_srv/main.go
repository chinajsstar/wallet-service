package main

import (
	"net/rpc"
	"../base/service"
	"../base/utils"
	"./handler"
	"fmt"
	"context"
	"time"
	l4g "github.com/alecthomas/log4go"
	"../push_srv/db"
)

const PushSrvConfig = "push.json"

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	l4g.LoadConfiguration(appDir + "/log.xml")
	defer l4g.Close()

	cfgPath := appDir + "/" + PushSrvConfig
	db.Init(cfgPath)

	handler.PushInstance().Init()

	// create service node
	fmt.Println("config path:", cfgPath)
	nodeInstance, err := service.NewServiceNode(cfgPath)
	if nodeInstance == nil || err != nil{
		l4g.Error("Create service node failed: %s", err.Error())
		return
	}
	rpc.Register(nodeInstance)

	// register apis
	service.RegisterNodeApi(nodeInstance, handler.PushInstance())

	// start service node
	ctx, cancel := context.WithCancel(context.Background())
	service.StartNode(ctx, nodeInstance)

	time.Sleep(time.Second*1)
	for ; ;  {
		fmt.Println("Input 'quit' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "quit" {
			cancel()
			break;
		}
	}

	l4g.Info("Waiting all routine quit...")
	service.StopNode(nodeInstance)
	l4g.Info("All routine is quit...")
}