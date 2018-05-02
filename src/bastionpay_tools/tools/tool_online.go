package tools

import (
	"fmt"
	"log"
	"blockchain_server/service"
	"blockchain_server/types"
	"bastionpay_tools/handler"
	"errors"
	"strconv"
	"bastionpay_tools/function"
	"strings"
	"bastionpay_tools/httphandler"
)

type OnLine struct{
	*function.Functions
	isStart bool
}

func (ol *OnLine) Usage() {
	fmt.Println("Usage: ")
	fmt.Println(">loadonlineaddress uniname")
	fmt.Println("	load online format address(parameter：uniname)")
	fmt.Println(">buildtxcmd type chiperprikey form to value txfilepath")
	fmt.Println("	build a test tx(parameter：type chiperprikey form to value txfilepath)")
	fmt.Println(">sendsignedtx txsignedfilepath")
	fmt.Println("	sendsignedtx(parameter: txsignedfilepath)")
	fmt.Println(">downloadfile url")
	fmt.Println("	download file from server(parameter：url)")
	fmt.Println(">uploadfile filepath url")
	fmt.Println("	upload file to server(parameter：url)")
}

func (ol *OnLine) Init(clientManager *service.ClientManager, dataDir string) error {
	fmt.Println("================================")
	fmt.Println("BastionPay online tool")
	fmt.Println("================================")

	ol.isStart = false

	// functions
	ol.Functions = &function.Functions{}
	ol.Functions.Init(clientManager, dataDir)

	// usage
	ol.Usage()
	return nil
}

func (ol *OnLine) Execute(argv []string) (string, error) {
	var err error
	var res string
	if argv[0] == "loadonlineaddress" {
		if len(argv) != 2 {
			log.Println("format：loadonlineaddress uniname")
			return "", errors.New("command error")
		}

		uniName := argv[1]
		accs, err := ol.LoadOnlineAddress(uniName)
		if err != nil {
			log.Println("loadonlineaddress failed: ", err.Error())
			return "", err
		}
		len := len(accs)
		for i := 0; i < len && i < 5; i++ {
			fmt.Println("index: ", i)
			fmt.Println("address: ", accs[i].Address)
			fmt.Println("prikey: ", accs[i].PrivateKey)
		}
		for i := len-5; i>=0 && i < len; i++ {
			fmt.Println("index: ", i)
			fmt.Println("address: ", accs[i].Address)
			fmt.Println("prikey: ", accs[i].PrivateKey)
		}
	} else if argv[0] == "buildtxcmd" {
		err = BuildTxTest(ol.GetClientManager(), argv)
	}else if argv[0] == "sendsignedtx" {
		if len(argv) != 2 {
			log.Println("format：sendsignedtx txsingedfilepath")
			return "", errors.New("command is error")
		}

		if ol.isStart == false{
			ol.GetClientManager().Start()
			ol.isStart = true
		}

		txSignedFilePath := argv[1]
		err = ol.SendSignedTx(txSignedFilePath)
	}else if argv[0] == "downloadfile" {
		if len(argv) != 2 {
			log.Println("format：downloadfile url")
			return "", errors.New("command is error")
		}

		url := argv[1]
		pps := strings.Split(url, "/")
		if len(pps) == 0 {
			log.Println("url error")
			return "", errors.New("url error")
		}

		filePath := ol.GetDataDir() + "/" + pps[len(pps)-1]
		err = httphandler.DownloadFile(filePath, url)
		fmt.Println("download fin: ", err)
	}else if argv[0] == "uploadfile" {
		if len(argv) != 3 {
			log.Println("format：uploadfile filepath url")
			return "", errors.New("command is error")
		}

		filePath := argv[1]
		url := argv[2]
		res, err = httphandler.UploadFile(filePath, url)
		fmt.Println("upload fin: ", err)
		fmt.Println(res)
	}else{
		ol.Usage()
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
	if txCmd == nil{
		fmt.Printf("创建交易失败\n")
		return errors.New("create tx failed")
	}
	txCmd.Tx.From = from
	txCmd.Chiperkey = chiperprikey

	var txArr []*types.CmdSendTx
	txArr = append(txArr, txCmd)

	txFilePath := argv[6]
	return handler.BuildTx(clientManager, txArr, txFilePath)
}