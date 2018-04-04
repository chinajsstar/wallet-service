package main

import (
	"net/rpc"
	"../base/service"
	"../data"
	"./handler"
	"fmt"
	"context"
	"time"
	"sync"
	"strings"
	"../base/config"
	"../base/utils"
	"strconv"
)

const ArithSrvConfig = "node.json"
var g_apisMap = make(map[string]service.CallNodeApi)

// 注册方法
func callArithFunction(req *data.SrvRequestData, res *data.SrvResponseData){
	h := g_apisMap[strings.ToLower(req.Data.Method.Function)]
	if h != nil {
		h(req, res)
	}else{
		res.Data.Err = data.ErrSrvInternalErr
		res.Data.ErrMsg = data.ErrSrvInternalErrText
	}

	//b1,_ := json.Marshal(*req)
	//fmt.Println("callNodeApi req: ", string(b1))
	//b,_ := json.Marshal(*res)
	//fmt.Println("callNodeApi ack: ", string(b))
}

func testPush(node *service.ServiceNode)  {
	for i := 0; i < 50; i++ {
		time.Sleep(time.Second*5)

		pData := data.UserResponseData{}
		pData.Method.Version = "v1"
		pData.Method.Srv = "arith"
		pData.Method.Function = "sub"
		pData.Value.LicenseKey = "e85d1e4f-5190-4edd-b459-a4b1a4b86764"
		pData.Value.Message = "abcd=" + strconv.Itoa(i)

		res := data.UserResponseData{}
		node.Push(&pData, &res)

		fmt.Println(res)
	}
}

func main() {
	var err error
	cn := config.ConfigNode{}
	if err = cn.Load(utils.GetRunDir()+"/config/"+ArithSrvConfig); err != nil{
		err = cn.Load(utils.GetCurrentDir() + "/config/" + ArithSrvConfig)
	}
	if err != nil {
		return
	}
	fmt.Println("config:", cn)

	wg := &sync.WaitGroup{}

	// 创建节点
	nodeInstance, _:= service.NewServiceNode(cn.SrvName, cn.SrvVersion)
	nodeInstance.RegisterData.Addr = cn.SrvAddr
	arith := new(handler.Arith)
	arith.RegisterApi(&nodeInstance.RegisterData.Functions, &g_apisMap)
	nodeInstance.Handler = callArithFunction

	nodeInstance.ServiceCenterAddr = cn.CenterAddr
	rpc.Register(nodeInstance)

	// 启动节点服务
	ctx, cancel := context.WithCancel(context.Background())
	nodeInstance.Start(ctx, wg)

	// 启动推送
	go testPush(nodeInstance)

	time.Sleep(time.Second*2)
	for ; ;  {
		fmt.Println("Input 'quit' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "quit" {
			cancel()
			break;
		}
	}

	fmt.Println("Waiting all routine quit...")
	wg.Wait()
	fmt.Println("All routine is quit...")
}
