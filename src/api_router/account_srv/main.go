package main

import (
	"net/rpc"
	"../base/service"
	"../data"
	"./handler"
	"./db"
	"fmt"
	"context"
	"time"
	"strings"
	"./install"
	"encoding/json"
	"../base/utils"
	"errors"
	"./user"
	"os"
)

const AccountSrvConfig = "account.json"

func installWallet(dir string) error {
	var err error

	fi, err := os.Open(dir+"/wallet.install")
	if err == nil {
		defer fi.Close()
		fmt.Println("Wallet is installed!!!")
		return nil
	}

	fmt.Println("Wallet is installing...")

	newRsa := false
	fmt.Println("1. create wallet rsa key...")
	_, err = os.Open(dir+"/private.pem")
	if err != nil {
		newRsa = true
	}
	_, err = os.Open(dir+"/public.pem")
	if err != nil {
		newRsa = true
	}
	if newRsa{
		pri := fmt.Sprintf(dir+"/private.pem")
		pub := fmt.Sprintf(dir+"/public.pem")
		err = utils.RsaGen(2048, pri, pub)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("create wallet rsa key...done")
	}else{
		fmt.Println("create wallet rsa key...exist")
	}

	fmt.Println("2. create wallet web admin...")
	uc, err := install.AddUser()
	if err != nil {
		fmt.Println("#error:",err)
		return err
	}
	b, _ := json.Marshal(*uc)

	var req data.SrvRequestData
	var res data.SrvResponseData
	req.Data.Argv.Message = string(b)

	handler.AccountInstance().Create(&req, &res)
	if res.Data.Err != data.NoErr {
		fmt.Println("#error:", res.Data.ErrMsg)
		return errors.New("创建admin失败")
	}
	fmt.Println("create wallet web admin...done")

	uca := user.AckUserCreate{}
	err = json.Unmarshal([]byte(res.Data.Value.Message), &uca)

	fmt.Println("记录license key:", uca.LicenseKey)

	fo, err := os.Create(dir+"/wallet.install")
	if err != nil {
		fmt.Println("#error:",err)
		return err
	}
	defer fo.Close()
	return nil
}

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	accountDir := appDir + "/account"
	err := os.MkdirAll(accountDir, os.ModePerm)
	if err!=nil && os.IsExist(err)==false {
		fmt.Println("#Error create account dir：", accountDir, "--", err)
		return
	}

	// start db
	db.Init()

	err = installWallet(accountDir)
	if err != nil {
		fmt.Println("#Error install：", err)
		return
	}

	// init
	handler.AccountInstance().Init(accountDir)

	// create service node
	cfgPath := appDir + "/" + AccountSrvConfig
	fmt.Println("config path:", cfgPath)
	nodeInstance, err := service.NewServiceNode(cfgPath)
	if nodeInstance == nil || err != nil{
		fmt.Println("#create service node failed:", err)
		return
	}
	rpc.Register(nodeInstance)

	// register APIs
	service.RegisterNodeApi(nodeInstance, handler.AccountInstance())

	// start service node
	ctx, cancel := context.WithCancel(context.Background())
	service.StartNode(ctx, nodeInstance)

	time.Sleep(time.Second*2)
	for ; ;  {
		fmt.Println("Input 'quit' to quit...")
		fmt.Println("Input 'createadmin' to create a user...")
		fmt.Println("Input 'loginadmin' to test the user...")
		var input string
		input = utils.ScanLine()

		argv := strings.Split(input, " ")

		if argv[0] == "quit" {
			cancel()
			break;
		}else if argv[0] == "createadmin" {
			uc, err := install.AddUser()
			if err != nil {
				fmt.Println("失败，",err)
				continue
			}
			b, _ := json.Marshal(*uc)

			var req data.SrvRequestData
			var res data.SrvResponseData
			req.Data.Argv.Message = string(b)

			handler.AccountInstance().Create(&req, &res)
			fmt.Println("createadmin err:", req)
			fmt.Println("createadmin ack:", res)
		}else if argv[0] == "loginadmin" {
			ul, err := install.LoginUser()
			if err != nil {
				fmt.Println("失败", err)
				continue
			}
			b, _ := json.Marshal(*ul)

			var req data.SrvRequestData
			var res data.SrvResponseData
			req.Data.Argv.Message = string(b)

			handler.AccountInstance().Login(&req, &res)
			fmt.Println("loginadmin err:", req)
			fmt.Println("loginadmin ack:", res)
		}
	}

	fmt.Println("Waiting all routine quit...")
	service.StopNode(nodeInstance)
	fmt.Println("All routine is quit...")
}