package main

import (
	"fmt"
	"sync/atomic"
	"time"
	"io/ioutil"
	"bastionpay_api/utils"
	"strings"
	"strconv"
	"api_router/base/config"
	"bastionpay_api/gateway"
)

const httpaddrGateway = "http://127.0.0.1:8082"
const httpaddrNigix = "http://127.0.0.1:8070"

func setGateway() error {
	var err error

	runDir, _ := utils.GetRunDir()
	cfgDir := config.GetBastionPayConfigDir()

	accountDir := cfgDir + "/" + config.BastionPayAccountDirName
	bastionPayPubkey, err := ioutil.ReadFile(accountDir + "/" + config.BastionPayPublicKey)
	if err != nil {
		return err
	}

	userKey := "1c75c668-f1ab-474b-9dae-9ed7950604b4"

	adminPubkey, err := ioutil.ReadFile(runDir + "/public_administrator.pem")
	if err != nil {
		return err
	}

	adminPrivkey, err := ioutil.ReadFile(runDir + "/private_administrator.pem")
	if err != nil {
		return err
	}

	gateway.SetBastionPaySetting(httpaddrGateway, bastionPayPubkey)
	gateway.SetUserSetting(userKey, adminPubkey, adminPrivkey)

	return nil
}

const(
	testMessage = "{\"a\":1, \"b\":1}"
	testApi = "/v1/arith/add"
)

var timeBegin,timeEnd time.Time

func DoApi(count *int64, right *int64, times int64){
	d, err := gateway.Output("/api" + testApi, []byte(testMessage))

	atomic.AddInt64(count, 1)
	if  err == nil{
		atomic.AddInt64(right, 1)
	}else{
		fmt.Println("#err:", string(d))
	}

	if atomic.CompareAndSwapInt64(count, times, times) {
		cost := time.Now().Sub(timeBegin)
		fmt.Println("结束时间：", time.Now())
		fmt.Println("finish...", *count, "...right...", *right, "...cost...", cost)
	}
}

func DoApiTest(count *int64, right *int64, times int64){
	d, err := gateway.Output("/apitest" + testApi, []byte(testMessage))

	atomic.AddInt64(count, 1)
	if  err == nil{
		atomic.AddInt64(right, 1)
	}else{
		fmt.Println("#err:", string(d))
	}

	if atomic.CompareAndSwapInt64(count, times, times) {
		cost := time.Now().Sub(timeBegin)
		fmt.Println("结束时间：", time.Now())
		fmt.Println("finish...", *count, "...right...", *right, "...cost...", cost)
	}
}

func main() {
	// 目录
	curDir, _ := utils.GetCurrentDir()
	runDir, _ := utils.GetRunDir()
	fmt.Println("当前目录：", curDir)
	fmt.Println("执行目录：", runDir)

	// load gateway
	err := setGateway()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 测试次数
	var runcounts = 100
	var count, right int64
	count = 0
	right = 0

	for ; ;  {
		fmt.Println("Please input command: ")
		var input string
		input = utils.ScanLine()

		argv := strings.Split(input, " ")

		count = 0
		right = 0
		timeBegin = time.Now();
		fmt.Println("开始时间：", timeBegin)

		if argv[0] == "q" {
			fmt.Println("I do quit")
			break;
		}else if argv[0] == "api" {
			d, err := gateway.Output("/api" + testApi, []byte(testMessage))
			fmt.Println("err==", err)
			fmt.Println("ack==", string(d))

		}else if argv[0] == "apitest" {
			d, err := gateway.Output("/apitest" + testApi, []byte(testMessage))
			fmt.Println("err==", err)
			fmt.Println("ack==", string(d))

		}else if argv[0] == "batchapi" {
			runcounts = 100
			if len(argv) > 1 {
				runcounts, _ = strconv.Atoi(argv[1])
			}

			for i := 0; i < runcounts; i++ {
				go DoApi(&count, &right, int64(runcounts))
			}
		} else if argv[0] == "batchapitest" {
			runcounts = 100
			if len(argv) > 1 {
				runcounts, _ = strconv.Atoi(argv[1])
			}

			for i := 0; i < runcounts; i++ {
				go DoApiTest(&count, &right, int64(runcounts))
			}
		}
	}
}