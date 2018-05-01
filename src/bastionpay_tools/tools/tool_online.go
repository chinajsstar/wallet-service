package tools

import (
	"fmt"
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	"bastionpay_tools/handler"
	"errors"
	"strconv"
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
		if len(argv) != 2 {
			fmt.Println("正确格式：loadonlineaddress 唯一标示")
			return "", errors.New("command is error")
		}

		uniName := argv[1]
		accs, err := handler.LoadOnlineAddress(ol.dataDir, uniName)
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
		err = BuildTxTest(ol.clientManager, argv)
	}else if argv[0] == "sendsignedtx" {
		if len(argv) != 2 {
			fmt.Println("正确格式：sendsignedtx 签名交易文件路径")
			return "", errors.New("command is error")
		}

		if ol.isStart == false{
			ol.clientManager.Start()
			ol.isStart = true
		}

		txSignedFilePath := argv[1]
		err = handler.SendSignedTx(ol.clientManager, txSignedFilePath)
	}else{
		usageonline()
		err = errors.New("unknown command")
	}

	return res, err
}

func BuildTxTest(clientManager *service.ClientManager, argv []string) (error) {
	if len(argv) != 7 {
		fmt.Println("正确格式：buildtxcmd 类型 加密私钥 从地址 去地址 数量 交易文件路径")
		return errors.New("command is error")
	}

	t := argv[1]
	chiperprikey := argv[2]
	from := argv[3]
	to := argv[4]
	value, err := strconv.Atoi(argv[5])
	if err != nil {
		fmt.Printf("数量不正确: %s\n", err.Error())
		return err
	}

	txCmd := service.NewSendTxCmd("message id", t, "", to, nil, uint64(value))
	txCmd.Tx.From = from
	txCmd.Chiperkey = chiperprikey

	var txArr []*types.CmdSendTx
	txArr = append(txArr, txCmd)

	txFilePath := argv[6]
	return handler.BuildTx(clientManager, txArr, txFilePath)
}