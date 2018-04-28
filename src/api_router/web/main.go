package main

import (
	"api_router/base/utils"
	"api_router/web/handler"
	"fmt"
	"time"
	l4g "github.com/alecthomas/log4go"
)

const WebSrvConfig = "web.json"

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	l4g.LoadConfiguration(appDir + "/log.xml")
	defer l4g.Close()

	// register apis
	web := new(handler.Web)
	if err := web.Init(appDir); err != nil{
		l4g.Error("Init service node failed: %s", err.Error())
		return
	}

	time.Sleep(time.Second*1)
	for ; ;  {
		fmt.Println("Input 'q' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "q" {
			break;
		}
	}

	l4g.Info("all routine quit...")
}
