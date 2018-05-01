package tools

import (
	"fmt"
	"blockchain_server/service"
	_ "github.com/mattn/go-sqlite3"
	"errors"
	"strconv"
	"bastionpay_tools/function"
	"log"
	"encoding/json"
)

type OffLine struct{
	*function.Functions
}

func (ol *OffLine)Usage()  {
	fmt.Println("Usage: ")
	fmt.Println(">newaddress type count")
	fmt.Println("	create new address(parameter：chaintype count)")
	fmt.Println(">loadonlineaddress uniname")
	fmt.Println("	load online format address(parameter：uniname)")
	fmt.Println(">loadofflineaddress uniname")
	fmt.Println("	load offline format address(parameter：uniname)")
	fmt.Println(">signtx txfilepath txsignedfilepath")
	fmt.Println("	sign transaction(parameter：txfilepath txsignedfilepath)")
}

func (ol *OffLine) Init(clientManager *service.ClientManager, dataDir string) error {
	fmt.Println("================================")
	fmt.Println("BastionPay offline tool")
	fmt.Println("================================")

	// functions
	ol.Functions = &function.Functions{}
	ol.Functions.Init(clientManager, dataDir)

	// usage
	ol.Usage()
	return nil
}

func (ol *OffLine)Execute(argv []string) (string, error) {
	var err error
	var res string
	fmt.Println(argv)
	if argv[0] == "newaddress" {
		if len(argv) != 3 {
			log.Println("format：newaddress type count")
			return "", errors.New("command error")
		}

		coinType := argv[1]
		count, err := strconv.Atoi(argv[2])
		if err != nil {
			log.Println("newaddress failed: ", err.Error())
			return "", err
		}

		uniNames, err := ol.NewAddress(coinType, uint32(count))
		bb, err := json.Marshal(uniNames)
		if err != nil {
			return "", err
		}
		res = string(bb)

		if err != nil {
			log.Println("newaddress failed: ", err.Error())
			return "", err
		}
	}else if argv[0] == "loadonlineaddress" {
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
	}else if argv[0] == "loadofflineaddress" {
		if len(argv) != 2 {
			log.Println("format：loadofflineaddress uniname")
			return "", errors.New("command error")
		}

		uniName := argv[1]
		accs, err := ol.LoadOfflineAddress(uniName)
		if err != nil {
			log.Println("loadofflineaddress failed: ", err.Error())
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
	}else if argv[0] == "signtx" {
		if len(argv) != 3 {
			log.Println("format: signtx txfilepath txsignedfilepath")
			return "", errors.New("command error")
		}

		txFilePath := argv[1]
		txSignedFilePath := argv[2]
		err = ol.SignTx(txFilePath, txSignedFilePath)
	}else{
		ol.Usage()
		err = errors.New("unknown command")
	}

	return res, err
}