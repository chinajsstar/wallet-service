package main

import (
	"api_router/web/handler"
	"fmt"
	"time"
	l4g "github.com/alecthomas/log4go"
	"api_router/base/config"
)

const WebSrvConfig = "web.json"

func main() {
	confDir := config.GetBastionPayConfigDir()

	l4g.LoadConfiguration(confDir + "/log.xml")
	defer l4g.Close()

	// register apis
	web := handler.NewWeb()
	if err := web.Init(confDir); err != nil{
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
