package tools

import (
	"fmt"
	"log"
	"blockchain_server/service"
	"blockchain_server/types"
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
	fmt.Println(">loadaddress uniaddressname")
	fmt.Println("	load format address(parameter：uniaddressname)")
	fmt.Println(">verifyaddressfile uniaddressname")
	fmt.Println("	verify address file md5(parameter：uniaddressname)")
	fmt.Println(">verifytxfile txpath")
	fmt.Println("	verifytxfile tx file md5(parameter：txpath)")
	fmt.Println(">buildtxcmd type chiperprikey form to value")
	fmt.Println("	build a test tx(parameter：type chiperprikey form to value)")
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

func (ol *OnLine) Execute(argv []string) (error) {
	var err error
	if argv[0] == "loadaddress" {
		if len(argv) != 2 {
			log.Println("format：loadaddress uniaddressname")
			return errors.New("command error")
		}

		dbName := argv[1]
		accs, err := ol.LoadAddress(dbName)
		if err != nil {
			log.Println("loadaddress failed: ", err.Error())
			return err
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
		log.Println("loadaddress fin: ", len)
	} else if argv[0] == "verifyaddressfile" {
		if len(argv) != 2 {
			log.Println("format：verifyaddressfile uniaddressname")
			return errors.New("command error")
		}

		dbName := argv[1]
		err = ol.VerifyAddressMd5(dbName)
		log.Println("verifyaddressfile fin: ", err)
	} else if argv[0] == "verifytxfile" {
		if len(argv) != 2 {
			log.Println("format：verifytxfile txpath")
			return errors.New("command error")
		}

		txPath := argv[1]
		err = ol.VerifyTxMd5(txPath)
		log.Println("verifytxfile fin: ", err)
	} else if argv[0] == "buildtxcmd" {
		res, err := BuildTxTest(ol, argv)
		log.Println("buildtxcmd fin: ", err)
		log.Println("buildtxcmd res: ", res)
	} else if argv[0] == "sendsignedtx" {
		if len(argv) != 2 {
			log.Println("format：sendsignedtx txsingedfilepath")
			return errors.New("command is error")
		}

		if ol.isStart == false{
			ol.GetClientManager().Start()
			ol.isStart = true
		}

		txSignedFilePath := argv[1]
		err = ol.SendSignedTx(txSignedFilePath)
		log.Println("sendsignedtx fin: ", err)
	}else if argv[0] == "downloadfile" {
		if len(argv) != 2 {
			log.Println("format：downloadfile url")
			return errors.New("command is error")
		}

		url := argv[1]
		pps := strings.Split(url, "/")
		if len(pps) == 0 {
			log.Println("url error")
			return errors.New("url error")
		}

		filePath := ol.GetDataDir() + "/" + pps[len(pps)-1]
		err = httphandler.DownloadFile(filePath, url)
		fmt.Println("download fin: ", err)
	}else if argv[0] == "uploadfile" {
		if len(argv) != 3 {
			log.Println("format：uploadfile filepath url")
			return errors.New("command is error")
		}

		filePath := argv[1]
		url := argv[2]
		res, err := httphandler.UploadFile(filePath, url)
		fmt.Println("upload fin: ", err)
		fmt.Println("res:", res)
	}else{
		ol.Usage()
		err = errors.New("unknown command")
	}

	return err
}

func BuildTxTest(ol *OnLine, argv []string) (string, error) {
	if len(argv) != 6 {
		fmt.Println("正确格式：buildtxcmd 类型 加密私钥 从地址 去地址 数量")
		return "", errors.New("command is error")
	}

	t := argv[1]
	chiperprikey := argv[2]
	from := argv[3]
	to := argv[4]
	value, err := strconv.Atoi(argv[5])
	if err != nil {
		fmt.Printf("数量不正确: %s\n", err.Error())
		return "", err
	}

	txCmd := service.NewSendTxCmd("message id", t, "", to, nil, uint64(value))
	if txCmd == nil{
		fmt.Printf("创建交易失败\n")
		return "", errors.New("create tx failed")
	}
	txCmd.Tx.From = from
	txCmd.Chiperkey = chiperprikey

	var txArr []*types.CmdSendTx
	txArr = append(txArr, txCmd)

	return ol.BuildTx(txArr)
}