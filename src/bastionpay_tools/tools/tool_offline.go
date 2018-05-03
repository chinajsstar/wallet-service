package tools

import (
	"fmt"
	"blockchain_server/service"
	_ "github.com/mattn/go-sqlite3"
	"errors"
	"strconv"
	"bastionpay_tools/function"
	"log"
	"crypto/md5"
	"encoding/hex"
	"bastionpay_tools/common"
)

type OffLine struct{
	*function.Functions
}

func (ol *OffLine)Usage()  {
	fmt.Println("Usage: ")
	fmt.Println(">newaddress type count saveddir")
	fmt.Println("	create new address(parameter：chaintype count saveddir)")
	fmt.Println(">loadaddress unidbname")
	fmt.Println("	load format address(parameter：unidbname)")
	fmt.Println(">checkmd5 unidbname")
	fmt.Println("	check address md5(parameter：unidbname)")
	fmt.Println(">testmd5 text")
	fmt.Println("	testmd5 text md5(parameter：text)")
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
			log.Println("format：loadaddress unidbname")
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
	} else if argv[0] == "checkmd5" {
		if len(argv) != 2 {
			log.Println("format：loadaddress unidbname")
			return errors.New("command error")
		}

		dbName := argv[1]
		err = ol.VerifyDbMd5(dbName)
		log.Println("checkmd5 fin: ", err)
	} else if argv[0] == "signtx" {
		if len(argv) != 3 {
			log.Println("format: signtx txfilepath txsignedfilepath")
			return errors.New("command error")
		}

		txFilePath := argv[1]
		txSignedFilePath := argv[2]
		err = ol.SignTx(txFilePath, txSignedFilePath)
		log.Println("checkmd5 fin: ", err)
	} else if argv[0] == "testmd5" {
		if len(argv) != 2 {
			log.Println("format: testmd5 text")
			return errors.New("command error")
		}
		func(){
			md5salt, err := common.GetSaltMd5HexByText(argv[1])
			fmt.Println("salt md5: ", md5salt)

			err = common.CompareSaltMd5HexByText(argv[1], md5salt)
			fmt.Println("compare salt md5: ", err)
		}()
		func(){
			h := md5.New()
			h.Write([]byte(argv[1]))
			sum := h.Sum(nil)

			md5salt := hex.EncodeToString(sum)
			fmt.Println("no salt md5: ", md5salt)

			err := common.CompareSaltMd5HexByText(argv[1], md5salt)
			fmt.Println("compare salt md5: ", err)
		}()
	}else{
		ol.Usage()
		err = errors.New("unknown command")
	}

	return err
}