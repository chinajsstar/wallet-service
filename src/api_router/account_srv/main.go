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
	"sync"
	"strings"
	"./install"
	"encoding/json"
	"../base/config"
	"../base/utils"
	"errors"
	"./user"
	"os"
)

const AccountSrvConfig = "node.json"
var g_apisMap = make(map[string]service.CallNodeApi)

// 注册方法
func callAuthFunction(req *data.SrvRequestData, res *data.SrvResponseData) {
	h := g_apisMap[strings.ToLower(req.Data.Function)]
	if h != nil {
		h(req, res)
	}else{
		res.Data.Err = data.ErrSrvInternalErr
		res.Data.ErrMsg = data.ErrSrvInternalErrText
	}

	fmt.Println("callNodeApi req: ", *req)
	fmt.Println("callNodeApi ack: ", *res)
}

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
	var err error
	var workerdir string

	cn := config.ConfigNode{}

	workerdir = utils.GetRunDir()
	if err = cn.Load(utils.GetRunDir()+"/config/"+AccountSrvConfig); err != nil{
		err = cn.Load(utils.GetCurrentDir() + "/config/" + AccountSrvConfig)
		workerdir = utils.GetCurrentDir()
	}
	if err != nil {
		return
	}
	fmt.Println("config:", cn)
	workerdir += "/worker"
	err = os.Mkdir(workerdir, os.ModePerm)
	if err!=nil && os.IsExist(err)==false {
		fmt.Println("#创建工作目录失败：", workerdir, "--", err)
		return
	}
	fmt.Println("workerdir:", workerdir)

	// 启动db
	db.Init()

	err = installWallet(workerdir)
	if err != nil {
		fmt.Println("安装失败：", err)
		return
	}

	wg := &sync.WaitGroup{}

	// 创建节点
	nodeInstance, _:= service.NewServiceNode(cn.SrvName, cn.SrvVersion)

	nodeInstance.RegisterData.Addr = cn.SrvAddr
	nodeInstance.Handler = callAuthFunction

	nodeInstance.ServiceCenterAddr = cn.CenterAddr

	// 注册API
	handler.AccountInstance().Init(workerdir)
	handler.AccountInstance().RegisterApi(&nodeInstance.RegisterData.Functions, &g_apisMap)

	rpc.Register(nodeInstance)

	// 启动节点服务
	ctx, cancel := context.WithCancel(context.Background())
	nodeInstance.Start(ctx, wg)

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
	wg.Wait()
	fmt.Println("All routine is quit...")
}