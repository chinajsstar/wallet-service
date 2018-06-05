package main

import (
	"fmt"
	"sync/atomic"
	"time"
	"io/ioutil"
	"bastionpay_api/utils"
	"strings"
	"strconv"
	"bastionpay_base/config"
	"bastionpay_api/gateway"
	"encoding/json"
	"reflect"
)

const httpaddrGateway = "http://127.0.0.1:8082"

func setGateway() error {
	var err error

	cfgDir := config.GetBastionPayConfigDir()

	accountDir := cfgDir + "/" + config.BastionPayAccountDirName
	bastionPayPubkey, err := ioutil.ReadFile(accountDir + "/" + config.BastionPayPublicKey)
	if err != nil {
		return err
	}

	userKey := "1c75c668-f1ab-474b-9dae-9ed7950604b4"

	adminPubkey, err := ioutil.ReadFile(cfgDir + "/public_administrator.pem")
	if err != nil {
		return err
	}

	adminPrivkey, err := ioutil.ReadFile(cfgDir + "/private_administrator.pem")
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
	var d []byte
	err := gateway.RunApi("/api" + testApi, []byte(testMessage), &d)

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
	var d []byte
	err := gateway.RunApiTest("/apitest" + testApi, []byte(testMessage), &d)

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

func checkMapIsValid(value interface{}, data map[string]interface{}) error {
	t := reflect.ValueOf(value).Type()
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("type is not struct")
	}

	keys := make(map[string]reflect.Type)
	n := t.NumField()
	for i := 0; i < n; i++ {
		jstr := t.Field(i).Tag.Get("json")
		if jstr == "-" {
			continue
		}
		if jstr == "" {
			jstr = t.Field(i).Name
		}
		keys[jstr] = t.Field(i).Type
	}

	for k, _ :=range data {
		_, ok := keys[k];
		if !ok {
			return fmt.Errorf("%s is error param", k)
		}

		//if ts != reflect.ValueOf(v).Type() {
		//	return fmt.Errorf("%s is error param type,%s-%s", k,ts.Name(),reflect.ValueOf(v).Type().Name())
		//}
	}

	return nil
}

type TT struct {
	Data string `json:"data"`
	T    int    `json:"t"`
	H    int    `json:"h"`
}

func test(msg string)  {
	var p map[string]interface{}
	json.Unmarshal([]byte(msg), &p)
	fmt.Println(p)

	err := checkMapIsValid(TT{}, p)
	fmt.Println(err)
}

func main() {
	//test("{\"data\":\"five\",\"t\":2}")
	//test("{\"data\":\"five\",\"h\":2}")
	//test("{\"data\":\"five\",\"g\":2}")
	//return

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
			var d []byte
			err := gateway.RunApi("/api" + testApi, []byte(testMessage), &d)
			fmt.Println("err==", err)
			fmt.Println("ack==", string(d))

		}else if argv[0] == "apitest" {
			var d []byte
			err := gateway.RunApiTest("/apitest" + testApi, []byte(testMessage), &d)
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