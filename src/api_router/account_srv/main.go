package main

import (
	"net/rpc"
	"../base/service"
	"../data"
	"./handler"
	"./db"
	"../base/utils"
	"fmt"
	"context"
	"time"
	"sync"
	"strings"
	"./install"
	"encoding/json"
)

const AccountSrvName = "account"
const AccountSrvVersion = "v1"
const (
	GateWayAddr = "127.0.0.1:8081"
	SrvAddr = "127.0.0.1:8092"
)
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

func main() {
	wg := &sync.WaitGroup{}

	// 启动db
	db.Init()

	// 创建节点
	nodeInstance, _:= service.NewServiceNode(AccountSrvName, AccountSrvVersion)

	nodeInstance.RegisterData.Addr = SrvAddr
	nodeInstance.Handler = callAuthFunction

	nodeInstance.ServiceCenterAddr = GateWayAddr

	// 注册API
	handler.AccountInstance().Init()
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
			b, _ := json.Marshal(uc)

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
			b, _ := json.Marshal(ul)

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