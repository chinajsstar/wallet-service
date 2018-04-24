package main

import (
	"../base/service"
	"fmt"
	"time"
	"context"
	"../base/utils"
	"net/rpc"
	l4g "github.com/alecthomas/log4go"
	"api_router/router/db"
	"api_router/router/handler"
)

const RouterConfig = "router.json"

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	l4g.LoadConfiguration(appDir + "/log.xml")
	defer l4g.Close()

	cfgPath := appDir + "/" + RouterConfig
	fmt.Println("config path:", cfgPath)

	db.Init(cfgPath)

	accountDir := appDir + "/account"
	handler.AuthInstance().Init(accountDir)

	// create service center
	router, err := service.NewRouter(cfgPath, handler.AuthInstance())
	if router == nil || err != nil {
		l4g.Error("Create service center failed: %s", err.Error())
		return
	}
	rpc.Register(router)

	// start service center
	ctx, cancel := context.WithCancel(context.Background())
	service.StartRouter(ctx, router)

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
	service.StopRouter(router)
	l4g.Info("All routine is quit...")
}

