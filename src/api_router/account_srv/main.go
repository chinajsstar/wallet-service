package main

import (
	"net/rpc"
	"../base/service"
	"../data"
	"./handler"
	"./db"
	"../utils"
	"fmt"
	"context"
	"time"
	"sync"
	"strings"
	"./install"
)

const AccountSrvName = "account"
const AccountSrvVersion = "v1"
const (
	GateWayAddr = "127.0.0.1:8081"
	SrvAddr = "127.0.0.1:8092"
)

// 注册方法
func callAuthFunction(req *data.ServiceCenterDispatchData, ack *data.ServiceCenterDispatchAckData){
	var err error
	err = func() error{
		var err error
		switch strings.ToLower(req.Function) {
		case "create":
			err = handler.AccountInstance().Create(req, ack)
			break
		case "login":
			err = handler.AccountInstance().Login(req, ack)
			break
		case "logout":
			err = handler.AccountInstance().Logout(req, ack)
			break
		case "updatepassword":
			err = handler.AccountInstance().UpdatePassword(req, ack)
			break
		case "listusers":
			err = handler.AccountInstance().ListUsers(req, ack)
			break
		}

		return err
	}()

	if err != nil {
		fmt.Println(err)
		ack.Err = data.ErrUserSrvRegister
		ack.ErrMsg = data.ErrUserSrvRegisterText
	}

	fmt.Println("callNodeApi req: ", *req)
	fmt.Println("callNodeApi ack: ", *ack)
}

func main() {
	wg := &sync.WaitGroup{}

	handler.AccountInstance().Init()

	// 启动db
	db.Init()

	// 创建节点
	nodeInstance, _:= service.NewServiceNode(AccountSrvName, AccountSrvVersion)
	nodeInstance.RegisterData.Addr = SrvAddr
	nodeInstance.RegisterData.RegisterFunction(new(handler.Account))
	nodeInstance.Handler = callAuthFunction

	nodeInstance.ServiceCenterAddr = GateWayAddr
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

			ack, err:= handler.AccountInstance().CreateWebAdmin(uc)
			fmt.Println("createadmin err:", err)
			fmt.Println("createadmin ack:", ack)
		}else if argv[0] == "loginadmin" {
			ul, err := install.LoginUser()
			if err != nil {
				fmt.Println("失败", err)
				continue
			}

			ack, err:= handler.AccountInstance().LoginWebAdmin(ul)
			fmt.Println("loginadmin err:", err)
			fmt.Println("loginadmin ack:", ack)
		}
	}

	fmt.Println("Waiting all routine quit...")
	wg.Wait()
	fmt.Println("All routine is quit...")
}