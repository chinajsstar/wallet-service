package main

import (
	"fmt"
	"strings"
	"api_router/base/utils"
	"os"
	"bastionpay_tools/tools"
)

func main()  {
	curDir, _ := utils.GetCurrentDir()
	dataDir := curDir + "/data"
	err := os.Mkdir(dataDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		fmt.Printf("创建数据目录失败：%s", err.Error())
		return
	}

	ol := &tools.OnLine{}
	err = ol.Start(dataDir)
	if err != nil {
		fmt.Printf("启动在线工具失败：%s", err.Error())
		return
	}

	for {
		var input string
		input = utils.ScanLine()
		argv := strings.Split(input, " ")

		if argv[0] == "q" {
			fmt.Println("I do quit")
			break;
		}else{
			ol.Execute(argv)
		}
	}
}