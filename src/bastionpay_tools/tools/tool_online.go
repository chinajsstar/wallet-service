package tools

import (
	"fmt"
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	"bastionpay_tools/handler"
	"errors"
)

func usageonline()  {
	fmt.Println("用法: ")
	fmt.Println(">q: 退出")
	fmt.Println(">loadonlineaddress uniname: 加载在线地址（参数：唯一标示）")
	fmt.Println(">buildtxcmd type chiperprikey form to value txfilepath: 生成交易（参数：类型 加密私钥 从地址 去地址 数量 交易文件路径）")
	fmt.Println(">sendsignedtx txsignedfilepath: 发送交易（参数：签名文件路径）")
}

type OnLine struct{
	clientManager *service.ClientManager
	dataDir string
	isStart bool
}

func (ol *OnLine)GetDataDir() string {
	return ol.dataDir
}

func (ol *OnLine) Start(dataDir string) error {
	fmt.Println("================================")
	fmt.Println("BastionPay在线工具")
	fmt.Println("================================")

	ol.dataDir = dataDir
	fmt.Printf("数据目录：%s", ol.dataDir)

	ol.clientManager = service.NewClientManager()
	ethClient, err := eth.ClientInstance()
	if nil != err {
		fmt.Printf("create client:%s error:%s", types.Chain_eth, err.Error())
		return err
	}
	// add client instance to manager
	ol.clientManager.AddClient(ethClient)

	usageonline()
	ol.isStart = false

	return nil
}

func (ol *OnLine) Execute(argv []string) (string, error) {
	var err error
	var res string
	if argv[0] == "loadonlineaddress" {
		accs, err := handler.LoadOnlineAddress(ol.dataDir, argv)
		if err != nil {
			fmt.Println("加载在线地址失败：", err.Error())
			return "", err
		}
		for i, acc := range accs {
			fmt.Println("索引：", i)
			fmt.Println("地址：", acc.Address)
			fmt.Println("私钥：", acc.PrivateKey)
		}
	} else if argv[0] == "buildtxcmd" {
		err = handler.BuildTxCmd(ol.clientManager, argv)
	}else if argv[0] == "sendsignedtx" {
		if ol.isStart == false{
			ol.clientManager.Start()
			ol.isStart = true
		}
		err = handler.SendSignedTx(ol.clientManager, argv)
	}else{
		usageonline()
		err = errors.New("unknown command")
	}

	return res, err
}