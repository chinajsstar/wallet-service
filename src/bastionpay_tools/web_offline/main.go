package main

import (
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	"api_router/base/utils"
	"bastionpay_tools/web_offline/handler"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"os"
	"bastionpay_tools/tools"
	"time"
	"strings"
	"api_router/base/config"
)

const(
	ConfigDirName = "BastionPay" 		// app dir
	DataDirName = "data"				// run dir

	ConfigFileName = "offline.json"		// config file
)
func main() {
	appDir, err:= utils.GetAppDir()
	if err != nil {
		fmt.Println("Get App directory failed: ", err)
		os.Exit(1)
	}
	runDir, err := utils.GetRunDir()
	if err != nil {
		fmt.Println("Get Run directory failed: ", err)
		os.Exit(1)
	}

	// 配置目录
	cfgDir := appDir + "/" + ConfigDirName
	l4g.Info("Config directory = %s", cfgDir)

	l4g.LoadConfiguration(cfgDir + "/log.xml")
	defer l4g.Close()

	configPath := cfgDir + "/" + ConfigFileName

	var port string
	err = config.LoadJsonNode(configPath, "port", &port)
	if err != nil {
		l4g.Crashf("", err)
	}

	// 数据目录
	dataDir := runDir + "/" + DataDirName
	err = os.Mkdir(dataDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		l4g.Crashf("Create data directory failed: %s", err.Error())
	}
	l4g.Info("Data directory = %s", dataDir)

	// 创建chain service
	clientManager := service.NewClientManager()
	// eth client
	ethClient, err := eth.ClientInstance()
	if nil != err {
		l4g.Crashf("Create client:%s error:%s", types.Chain_eth, err.Error())
	}
	clientManager.AddClient(ethClient)

	// 启动服务
	web := new(handler.Web)
	if err := web.Init(clientManager, dataDir); err != nil{
		l4g.Crashf("Init service failed: %s", err.Error())
	}
	if err := web.StartHttpServer(port); err != nil{
		l4g.Crashf("Start http service failed: %s", err.Error())
	}

	fmt.Println("Bastion pay offline tool started...")
	fmt.Println("Input 'q' to exit...")

	// debug mode
	// 启动工具
	ol := &tools.OffLine{}
	err = ol.Init(clientManager, dataDir)
	if err != nil {
		l4g.Crashf("Start Bastion offline tool failed: %s", err.Error())
	}
	time.Sleep(time.Second*1)
	for {
		var input string
		input = utils.ScanLine()
		argv := strings.Split(input, " ")

		if argv[0] == "q" {
			break;
		}else if argv[0] == "help" {
			ol.Usage()
		}else {
			ol.Execute(argv)
		}
	}
}
