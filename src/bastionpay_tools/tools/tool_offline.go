package tools

import (
	"fmt"
	"blockchain_server/service"
	_ "github.com/mattn/go-sqlite3"
	"errors"
	"strconv"
	"bastionpay_tools/function"
	"log"
)

type OffLine struct{
	*function.Functions
}

func (ol *OffLine)Usage()  {
	fmt.Println("Usage: ")
	fmt.Println(">newaddress type count saveddir")
	fmt.Println("	create new address(parameter：chaintype count saveddir)")
	fmt.Println(">loadaddress uniaddressname")
	fmt.Println("	load format address(parameter：uniaddressname)")
	fmt.Println(">verifyaddressfile uniaddressname")
	fmt.Println("	verify address file md5(parameter：uniaddressname)")
	fmt.Println(">verifytxfile txpath")
	fmt.Println("	verifytxfile tx file md5(parameter：txpath)")
	fmt.Println(">signtx txfilepath saveddir")
	fmt.Println("	sign transaction(parameter：txfilepath saveddir)")
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

func (ol *OffLine)Execute(argv []string) (error) {
	var err error
	fmt.Println(argv)
	if argv[0] == "newaddress" {
		if len(argv) != 4 {
			log.Println("format：newaddress type count saveddir")
			return errors.New("command error")
		}

		coinType := argv[1]
		count, err := strconv.Atoi(argv[2])
		if err != nil {
			log.Println("newaddress failed: ", err.Error())
			return err
		}

		savedDir := argv[3]
		uniName, err := ol.NewAddress(coinType, uint32(count), savedDir)
		if err != nil {
			log.Println("newaddress failed: ", err.Error())
			return err
		}

		if err != nil {
			log.Println("newaddress failed: ", err.Error())
			return err
		}
		log.Println("newaddress fin: ", uniName)
	}else if argv[0] == "loadaddress" {
		if len(argv) != 2 {
			log.Println("format：loadaddress uniaddressname")
			return errors.New("command error")
		}

		uniName := argv[1]
		accs, err := ol.LoadAddress(uniName)
		if err != nil {
			log.Println("loadonlineaddress failed: ", err.Error())
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
	} else if argv[0] == "signtx" {
		if len(argv) != 3 {
			log.Println("format: signtx txfilepath saveddir")
			return errors.New("command error")
		}

		txFilePath := argv[1]
		txSavedDir := argv[2]
		res, err := ol.SignTx(txFilePath, txSavedDir)
		log.Println("signtx fin: ", err)
		log.Println("signtx res: ", res)
	} else{
		ol.Usage()
		err = errors.New("unknown command")
	}

	return err
}