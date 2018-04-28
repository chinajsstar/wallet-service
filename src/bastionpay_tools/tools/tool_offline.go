package tools

import (
	"fmt"
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	_ "github.com/mattn/go-sqlite3"
	"bastionpay_tools/handler"
	"errors"
	"strconv"
)

func usageoffline()  {
	fmt.Println("用法: ")
	fmt.Println(">q: 退出")
	fmt.Println(">newaddress type count: 创建新地址（参数：类型 数量）")
	fmt.Println(">loadonlineaddress uniname: 加载在线地址（参数：唯一标示）")
	fmt.Println(">loadofflineaddress uniname: 加载离线地址（参数：唯一标示）")
	fmt.Println(">signtx txfilepath txsignedfilepath: 签名交易（参数：交易文件路径 签名交易路径）")
}

type OffLine struct{
	clientManager *service.ClientManager
	dataDir string
}

func (ol *OffLine)GetDataDir() string {
	return ol.dataDir
}

func (ol *OffLine) Start(dataDir string) error {
	fmt.Println("================================")
	fmt.Println("BastionPay离线工具")
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

	// send a tx
	//tmp_account_eth := &types.Account{
	//	"0x04e2b6c9bfeacd4880d99790a03a3db4ad8d87c82bb7d72711b277a9a03e49743077f3ae6d0d40e6bc04eceba67c2b3ec670b22b30d57f9d6c42779a05fba097536c412af73be02d1642aecea9fa7082db301e41d1c3c2686a6a21ca431e7e8605f761d8e12d61ca77605b31d707abc3f17bc4a28f4939f352f283a48ed77fc274b039590cc2c43ef739bd3ea13e491316",
	//	"0x54b2e44d40d3df64e38487dd4e145b3e6ae25927"}

	usageoffline()

	return nil
}

func (ol *OffLine)Execute(argv []string) (string, error) {
	var err error
	var res string
	if argv[0] == "newaddress" {
		if len(argv) != 3 {
			fmt.Println("正确格式：newaddress 类型 数量")
			return "", errors.New("command error")
		}

		coinType := argv[1]
		count, _ := strconv.Atoi(argv[2])

		res, err = handler.NewAddress(ol.clientManager, ol.dataDir, coinType, uint32(count))
	}else if argv[0] == "loadonlineaddress" {
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
	}else if argv[0] == "loadofflineaddress" {
		if len(argv) != 2 {
			fmt.Println("正确格式：loadofflineaddress 唯一标示")
			return "", errors.New("command is error")
		}

		uniName := argv[1]
		accs, err := handler.LoadOfflineAddress(ol.dataDir, uniName)
		if err != nil {
			fmt.Println("加载在线地址失败：", err.Error())
			return "", err
		}
		for i, acc := range accs {
			fmt.Println("索引：", i)
			fmt.Println("地址：", acc.Address)
			fmt.Println("私钥：", acc.PrivateKey)
		}
	}else if argv[0] == "signtx" {
		if len(argv) != 3 {
			fmt.Println("正确格式：signtx 交易文件路径 签名交易文件路径")
			return "", errors.New("command is error")
		}

		txFilePath := argv[1]
		txSignedFilePath := argv[2]

		err = handler.SignTx(ol.clientManager, ol.dataDir, txFilePath, txSignedFilePath)
	}else{
		usageoffline()
		err = errors.New("unknown command")
	}

	return res, err
}