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

	//tmp_account_ztoken := &types.Account{
	//	"0x04e2b6c9bfeacd4880d99790a03a3db4ad8d87c82bb7d72711b277a9a03e49743077f3ae6d0d40e6bc04eceba67c2b3ec670b22b30d57f9d6c42779a05fba097536c412af73be02d1642aecea9fa7082db301e41d1c3c2686a6a21ca431e7e8605f761d8e12d61ca77605b31d707abc3f17bc4a28f4939f352f283a48ed77fc274b039590cc2c43ef739bd3ea13e491316",
	//	"0x54b2e44d40d3df64e38487dd4e145b3e6ae25927"}

	t := argv[1]
	chiperprikey := argv[2]
	from := argv[3]
	to := argv[4]
	value, err := strconv.ParseFloat(argv[5], 64)
	if err != nil {
		fmt.Printf("数量不正确: %s\n", err.Error())
		return "", err
	}
	//msgId, coinName, fromKey, to, tkname, tokenFromkey string, value float64
	txCmd, err := service.NewSendTxCmd("", t, chiperprikey, to, "", "", value)
	//txCmd := service.NewSendTxCmd("", t, "", to, nil, uint64(value))
	if err != nil{
		fmt.Printf("创建交易失败:%s", err.Error())
		return "", err
	}
	txCmd.Tx.From = from
	txCmd.FromKey = chiperprikey

	var txArr []*types.CmdSendTx
	txArr = append(txArr, txCmd)

	return ol.BuildTx(txArr)
}