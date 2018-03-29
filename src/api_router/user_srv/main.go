package main

import (
	"net/rpc"
	"../base/service"
	"../data"
	"./handler"
	"./db"
	"./user"
	"../utils"
	"fmt"
	"context"
	"time"
	"sync"
	"strings"
	"crypto/md5"
	"io/ioutil"
	"encoding/base64"
)

const UserSrvName = "user"
const UserSrvVersion = "v1"
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
	nodeInstance, _:= service.NewServiceNode(UserSrvName, UserSrvVersion)
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
		var input string
		input = utils.ScanLine()

		argv := strings.Split(input, " ")

		if argv[0] == "quit" {
			cancel()
			break;
		}else if argv[0] == "createwebadmin" {
			if len(argv) < 6 {
				fmt.Println("username phone email password pubkey, ", len(argv))
				continue
			}

			pubPath := fmt.Sprintf("/Users/henly.liu/workspace/public_%s.pem", argv[5])
			pubKey, err := ioutil.ReadFile(pubPath)
			if err != nil {
				fmt.Println(err)
				continue
			}

			h := md5.New()
			h.Write([]byte(argv[4]))
			pw := h.Sum(nil)
			pwss := base64.StdEncoding.EncodeToString(pw)

			// name pubkey
			uc := user.UserCreate{}
			uc.UserName = argv[1]
			uc.Phone = argv[2]
			uc.Email = argv[3]
			uc.Password = pwss
			uc.Language = "ch"
			uc.Country = "China"
			uc.TimeZone = "Beijing"
			uc.GoogleAuth = ""
			uc.PublicKey = string(pubKey)

			ack, err:= handler.AccountInstance().CreateWebAdmin(&uc)
			fmt.Println("createwebadmin err:", err)
			fmt.Println("createwebadmin ack:", ack)
		}else if argv[0] == "loginwebadmin" {
			if len(argv) < 3 {
				fmt.Println("username password pubkey")
				continue
			}
			h := md5.New()
			h.Write([]byte(argv[2]))
			pw := h.Sum(nil)
			pwss := base64.StdEncoding.EncodeToString(pw)

			// name pubkey
			ul := user.UserLogin{}
			ul.UserName = argv[1]
			ul.Phone = "13585596201"
			ul.Email = "henly.liu@blockshine.com"
			ul.Password = pwss

			ack, err:= handler.AccountInstance().LoginWebAdmin(&ul)
			fmt.Println("loginwebadmin err:", err)
			fmt.Println("loginwebadmin ack:", ack)
		}
	}

	fmt.Println("Waiting all routine quit...")
	wg.Wait()
	fmt.Println("All routine is quit...")
}