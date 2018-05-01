package main

import (
	"api_router/base/utils"
	"bastionpay_tools/web_offline/handler"
	"fmt"
	"time"
	l4g "github.com/alecthomas/log4go"
	"os"
	"bastionpay_tools/tools"
	"strings"
)

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	l4g.LoadConfiguration(appDir + "/log.xml")
	defer l4g.Close()

	curDir, _ := utils.GetCurrentDir()
	dataDir := curDir + "/data"
	err := os.Mkdir(dataDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		fmt.Printf("创建数据目录失败：%s", err.Error())
		return
	}
	ol := &tools.OffLine{}
	err = ol.Start(dataDir)
	if err != nil {
		fmt.Printf("启动离线工具失败：%s", err.Error())
		return
	}

	web := new(handler.Web)
	if err := web.Init(ol); err != nil{
		l4g.Error("Init service node failed: %s", err.Error())
		return
	}

	time.Sleep(time.Second*1)
	for {
		var input string
		input = utils.ScanLine()
		argv := strings.Split(input, " ")

		if argv[0] == "q" {
			break;
		}else{
			ol.Execute(argv)
		}
	}

	l4g.Info("all routine quit...")
}
