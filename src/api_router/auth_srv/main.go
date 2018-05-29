package main

import (
	//"api_router/base/service"
	service "api_router/base/service2"
	"api_router/auth_srv/handler"
	"fmt"
	"context"
	"time"
	l4g "github.com/alecthomas/log4go"
	"api_router/auth_srv/db"
	"api_router/base/config"
	"bastionpay_api/utils"
)

const AuthSrvConfig = "auth.json"

func main() {
	cfgDir := config.GetBastionPayConfigDir()

	l4g.LoadConfiguration(cfgDir + "/log.xml")
	defer l4g.Close()

	defer utils.PanicPrint()

	cfgPath := cfgDir + "/" + AuthSrvConfig
	fmt.Println("config path:", cfgPath)
	db.Init(cfgPath)

	// create service node
	nodeInstance, err := service.NewServiceNode(cfgPath)
	if nodeInstance == nil || err != nil{
		l4g.Error("Create service node failed: %s", err.Error())
		return
	}

	accountDir := cfgDir + "/" + config.BastionPayAccountDirName
	handler.AuthInstance().Init(accountDir, nodeInstance)

	// register apis
	service.RegisterNodeApi(nodeInstance, handler.AuthInstance())

	// start service node
	ctx, cancel := context.WithCancel(context.Background())
	service.StartNode(ctx, nodeInstance)

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