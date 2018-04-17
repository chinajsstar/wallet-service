package main

import (
	"net/rpc"
	"../base/service"
	"../base/data"
	"./handler"
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
	l4g "github.com/alecthomas/log4go"
)

const AccountSrvConfig = "account.json"

func installWallet(dir string) error {
	var err error

	fi, err := os.Open(dir+"/wallet.install")
	if err == nil {
		defer fi.Close()
		l4g.Info("Super wallet is installed, have a fun...")
		return nil
	}

	l4g.Info("First time to use Super wallet, need to install step by step...")
	newRsa := false
	l4g.Info("1. Create wallet rsa key...")
	_, err = os.Open(dir+"/private.pem")
	if err != nil {
		newRsa = true
	}
	_, err = os.Open(dir+"/public.pem")
	if err != nil {
		newRsa = true
	}
	if newRsa{
		l4g.Info("Create new wallet rsa key in %s", dir)
		pri := fmt.Sprintf(dir+"/private.pem")
		pub := fmt.Sprintf(dir+"/public.pem")
		err = utils.RsaGen(2048, pri, pub)
		if err != nil {
			return err
		}
	}else{
		l4g.Info("A wallet rsa key is exist...")
	}

	l4g.Info("2. Create wallet genesis admin...")
	uc, err := install.AddUser()
	if err != nil {
		return err
	}
	b, _ := json.Marshal(*uc)

	var req data.SrvRequestData
	var res data.SrvResponseData
	req.Data.Argv.Message = string(b)
	handler.AccountInstance().Create(&req, &res)
	if res.Data.Err != data.NoErr {
		return errors.New(res.Data.ErrMsg)
	}

	uca := user.AckUserCreate{}
	err = json.Unmarshal([]byte(res.Data.Value.Message), &uca)

	l4g.Info("3. Record genesis admin license key: %s", uca.LicenseKey)
	l4g.Info("4. Record super wallet rsa pub key: %s", uca.ServerPublicKey)

	// write a tag file
	fo, err := os.Create(dir+"/wallet.install")
	if err != nil {
		return err
	}
	defer fo.Close()
	return nil
}

func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	l4g.LoadConfiguration(appDir + "/log.xml")
	defer l4g.Close()

	accountDir := appDir + "/account"
	err := os.MkdirAll(accountDir, os.ModePerm)
	if err!=nil && os.IsExist(err)==false {
		l4g.Error("Create dir failed：%s - %s", accountDir, err.Error())
		return
	}

	err = installWallet(accountDir)
	if err != nil {
		l4g.Error("Install super wallet failed: %s", err.Error())
		return
	}

	// init
	handler.AccountInstance().Init(accountDir)

	// create service node
	cfgPath := appDir + "/" + AccountSrvConfig
	l4g.Info("config path: %s", cfgPath)
	nodeInstance, err := service.NewServiceNode(cfgPath)
	if nodeInstance == nil || err != nil{
		l4g.Error("Create service node failed: %s", err.Error())
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
		fmt.Println("Input 'createuser' to create a user...")
		fmt.Println("Input 'loginuser' to test the user...")
		var input string
		input = utils.ScanLine()

		argv := strings.Split(input, " ")

		if argv[0] == "quit" {
			cancel()
			break;
		}else if argv[0] == "createuser" {
			uc, err := install.AddUser()
			if err != nil {
				l4g.Error("createuser failed: %s",err.Error())
				continue
			}
			b, _ := json.Marshal(*uc)

			var req data.SrvRequestData
			var res data.SrvResponseData
			req.Data.Argv.Message = string(b)

			handler.AccountInstance().Create(&req, &res)
			l4g.Info("createuser req:", req)
			l4g.Info("createuser res:", res)
		}else if argv[0] == "loginuser" {
			ul, err := install.LoginUser()
			if err != nil {
				l4g.Error("loginuser failed: %s",err.Error())
				continue
			}
			b, _ := json.Marshal(*ul)

			var req data.SrvRequestData
			var res data.SrvResponseData
			req.Data.Argv.Message = string(b)

			handler.AccountInstance().Login(&req, &res)
			l4g.Info("loginuser req:", req)
			l4g.Info("loginuser res:", res)
		}
	}

	l4g.Info("Waiting all routine quit...")
	service.StopNode(nodeInstance)
	l4g.Info("All routine is quit...")
}