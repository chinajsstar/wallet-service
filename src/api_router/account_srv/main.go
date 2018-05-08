package main

import (
	//"api_router/base/service"
	service "api_router/base/service2"
	"api_router/account_srv/handler"
	"fmt"
	"context"
	"time"
	"strings"
	"api_router/account_srv/install"
	"api_router/base/utils"
	"os"
	l4g "github.com/alecthomas/log4go"
	"api_router/account_srv/db"
	"api_router/base/config"
)

const AccountSrvConfig = "account.json"

func main() {
	cfgDir := config.GetBastionPayConfigDir()

	l4g.LoadConfiguration(cfgDir + "/log.xml")
	defer l4g.Close()

	cfgPath := cfgDir + "/" + AccountSrvConfig
	db.Init(cfgPath)

	accountDir := cfgDir + "/" + config.BastionPayAccountDirName
	err := os.MkdirAll(accountDir, os.ModePerm)
	if err != nil && os.IsExist(err) == false {
		l4g.Error("Create dir failed：%s - %s", accountDir, err.Error())
		return
	}

	err = install.InstallBastionPay(accountDir)
	if err != nil {
		l4g.Error("Install super wallet failed: %s", err.Error())
		return
	}

	// init
	handler.AccountInstance().Init(accountDir)

	// create service node
	l4g.Info("config path: %s", cfgPath)
	nodeInstance, err := service.NewServiceNode(cfgPath)
	if nodeInstance == nil || err != nil{
		l4g.Error("Create service node failed: %s", err.Error())
		return
	}

	// register APIs
	service.RegisterNodeApi(nodeInstance, handler.AccountInstance())

	// start service node
	ctx, cancel := context.WithCancel(context.Background())
	service.StartNode(ctx, nodeInstance)

	time.Sleep(time.Second*2)
	for ; ;  {
		fmt.Println("Input 'q' to quit...")
		var input string
		input = utils.ScanLine()

		argv := strings.Split(input, " ")

		if argv[0] == "q" {
			cancel()
			break;
		}
	}

	l4g.Info("Waiting all routine quit...")
	service.StopNode(nodeInstance)
	l4g.Info("All routine is quit...")
}