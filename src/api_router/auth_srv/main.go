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
	"io/ioutil"
	"crypto/sha512"
	"crypto"
	"strings"
	"sync/atomic"
	"../base/config"
	"../base/utils"
)

const AuthSrvConfig = "node.json"

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

	//fmt.Println("callNodeApi req: ", *req)
	//fmt.Println("callNodeApi ack: ", *res)
}

func main() {
	var err error
	var workerdir string

	cn := config.ConfigNode{}

	workerdir = utils.GetRunDir()
	if err = cn.Load(utils.GetRunDir()+"/config/"+AuthSrvConfig); err != nil{
		err = cn.Load(utils.GetCurrentDir() + "/config/" + AuthSrvConfig)

		workerdir = utils.GetCurrentDir()
	}
	if err != nil {
		return
	}
	fmt.Println("config:", cn)

	workerdir += "/worker"
	handler.AuthInstance().Init(workerdir)

	wg := &sync.WaitGroup{}

	// 启动db
	db.Init()

	// 创建节点
	nodeInstance, _:= service.NewServiceNode(cn.SrvName, cn.SrvVersion)
	nodeInstance.RegisterData.Addr = cn.SrvAddr
	handler.AuthInstance().RegisterApi(&nodeInstance.RegisterData.Functions, &g_apisMap)
	nodeInstance.Handler = callAuthFunction

	nodeInstance.ServiceCenterAddr = cn.CenterAddr
	rpc.Register(nodeInstance)

	// 启动节点服务
	ctx, cancel := context.WithCancel(context.Background())
	nodeInstance.Start(ctx, wg)

	var cipherData []byte

	time.Sleep(time.Second*2)
	for ; ;  {
		fmt.Println("Input 'quit' to quit...")
		var input string
		fmt.Scanln(&input)

		if input == "quit" {
			cancel()
			break;
		}else if input == "rsatest" {
			var priKey, pubKey []byte
			priKey, err = ioutil.ReadFile("/Users/henly.liu/workspace/private_wallet.pem")
			if err != nil {
				fmt.Println(err)
				continue
			}
			pubKey, err = ioutil.ReadFile("/Users/henly.liu/workspace/public_wallet.pem")
			if err != nil {
				fmt.Println(err)
				continue
			}

			var data []byte
			for i := 0; i < 123; i++ {
				data = append(data, byte(i))
			}

			// 测试次数
			timeBegin := time.Now()

			var runcounts int
			var count, right int64
			runcounts = 1
			count = 0
			right = 0

			testfunc := func(count *int64, right *int64, runcounts int64) {
				fmt.Println("原始数据：", len(data))
				fmt.Println(data)
				// en
				cipherData, err = utils.RsaEncrypt(data, pubKey, utils.RsaEncodeLimit2048)
				if err != nil {
					fmt.Println(err)
				}

				fmt.Println("加密后数据：", len(cipherData))
				fmt.Println(cipherData)

				// de
				var originData []byte
				originData, err = utils.RsaDecrypt(cipherData, priKey, utils.RsaDecodeLimit2048)
				if err != nil {
					fmt.Println(err)
				}

				fmt.Println("解密后数据：")
				fmt.Println(originData)

				atomic.AddInt64(count, 1)
				if  err == nil{
					atomic.AddInt64(right, 1)
				}else{
					fmt.Println("#err:")
				}

				if atomic.CompareAndSwapInt64(count, runcounts, runcounts) {
					cost := time.Now().Sub(timeBegin)
					fmt.Println("结束时间：", time.Now())
					fmt.Println("finish...", *count, "...right...", *right, "...cost...", cost)
				}
			}

			for i := 0; i < runcounts; i++ {
				go testfunc(&count, &right, int64(runcounts))
				//testfunc(&count, &right, int64(runcounts))
			}

			cost := time.Now().Sub(timeBegin)
			fmt.Println("结束时间：", time.Now())
			fmt.Println("finish...", count, "...right...", right, "...cost...", cost)

		}else if input == "rsatest2" {
			var priKey, pubKey []byte
			priKey, err = ioutil.ReadFile("/Users/henly.liu/workspace/private.pem")
			if err != nil {
				fmt.Println(err)
				continue
			}
			pubKey, err = ioutil.ReadFile("/Users/henly.liu/workspace/public.pem")
			if err != nil {
				fmt.Println(err)
				continue
			}

			// sign
			var hashData []byte
			hs := sha512.New()
			hs.Write(cipherData)
			hashData = hs.Sum(nil)
			fmt.Println("哈希后数据：")
			fmt.Println(hashData)

			var signData []byte
			signData, err = utils.RsaSign(crypto.SHA512, hashData, priKey)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("签名后数据：")
			fmt.Println(signData)

			// verify
			err = utils.RsaVerify(crypto.SHA512, hashData, signData, pubKey)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("验证后数据：")
			fmt.Println(err)
		}
	}

	fmt.Println("Waiting all routine quit...")
	wg.Wait()
	fmt.Println("All routine is quit...")
}