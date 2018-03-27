package main

import (
	"../base/nethelper"
	"../data"
	"fmt"
	"encoding/json" // for json get
	"sync/atomic"
	"net/rpc"
	"log"
	"time"
	"io/ioutil"
	"../utils"
	"crypto/sha512"
	"crypto"
	"encoding/base64"
)

var timeBegin,timeEnd time.Time

func DoTest(params interface{}, count *int64, right *int64, times int64){
	ackData := data.ServiceCenterDispatchAckData{}
	err := nethelper.CallJRPCToHttpServer("127.0.0.1:8080", "/wallet", data.MethodServiceCenterDispatch, params, &ackData)

	atomic.AddInt64(count, 1)
	if  err == nil && ackData.Err==0{
		atomic.AddInt64(right, 1)
	}

	if atomic.CompareAndSwapInt64(count, times, times) {
		cost := time.Now().Sub(timeBegin)
		fmt.Println("结束时间：", time.Now())
		fmt.Println("finish...", *count, "...right...", *right, "...cost...", cost)
	}
}

func DoTest2(client *rpc.Client, params interface{}, count *int64, right *int64, times int64){
	ackData := data.ServiceCenterDispatchAckData{}
	err := nethelper.CallJRPCToHttpServerOnClient(client, data.MethodServiceCenterDispatch, params, &ackData)

	atomic.AddInt64(count, 1)
	if  err == nil && ackData.Err==0{
		atomic.AddInt64(right, 1)
	}

	if atomic.CompareAndSwapInt64(count, times, times) {
		cost := time.Now().Sub(timeBegin)
		fmt.Println("结束时间：", time.Now())
		fmt.Println("finish...", *count, "...right...", *right, "...cost...", cost)
	}
}

func DoTestTcp(params interface{}, count *int64, right *int64, times int64){
	ackData := data.ServiceCenterDispatchAckData{}

	err := nethelper.CallJRPCToTcpServer("127.0.0.1:8090", data.MethodServiceNodeCall, params, &ackData)

	atomic.AddInt64(count, 1)
	if  err == nil && ackData.Err==0{
		atomic.AddInt64(right, 1)
	}

	if atomic.CompareAndSwapInt64(count, times, times) {
		cost := time.Now().Sub(timeBegin)
		fmt.Println("结束时间：", time.Now())
		fmt.Println("finish...", *count, "...right...", *right, "...cost...", cost)
	}
}

func DoTestTcp2(client *rpc.Client, params interface{}, count *int64, right *int64, times int64){
	ackData := data.ServiceCenterDispatchAckData{}
	err := nethelper.CallJRPCToTcpServerOnClient(client, data.MethodServiceNodeCall, params, &ackData)

	atomic.AddInt64(count, 1)
	if  err == nil && ackData.Err==0{
		atomic.AddInt64(right, 1)
	}

	if atomic.CompareAndSwapInt64(count, times, times) {
		cost := time.Now().Sub(timeBegin)
		fmt.Println("结束时间：", time.Now())
		fmt.Println("finish...", *count, "...right...", *right, "...cost...", cost)
	}
}

// http rpc风格
// curl -d '{"method":"ServiceCenter.Dispatch", "params":[{"srv":"v1.arith", "function":"add","argv":"{\"a\":\"hello, \", \"b\":\"world\"}"}], "id": 1}' http://localhost:8080/rpc
// curl -d '{
// "method":"ServiceCenter.Dispatch",
// "params":[{"srv":"v1.arith", "function":"add", "argv":"{\"a\":\"hello, \", \"b\":\"world\"}}],
// "id": 1
// }'
// http://localhost:8080/rpc

// http restful风格
// curl -d '{"argv":"{\"a\":2, \"b\":1}"}' http://localhost:8080/restful/v1/arith/add
func main() {

	var err error
	index := 1
	priPath := fmt.Sprintf("/Users/henly.liu/workspace/private_%d.pem", index)
	serverPubPath := fmt.Sprintf("/Users/henly.liu/workspace/public.pem")
	licenseKey := fmt.Sprintf("licensekey_%d", index)

	var priKey, serverPubKey []byte
	priKey, err = ioutil.ReadFile(priPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	serverPubKey, err = ioutil.ReadFile(serverPubPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	const times = 100;
	var count, right int64
	count = 0
	right = 0

	yourmessage := "[{\"a\":1, \"b\":2}]"
	yourlicensekey := licenseKey

	// 用户数据
	var ud data.UserData
	ud.LicenseKey = yourlicensekey
	ud.Message = func() string{
		// 用我们的pub加密message ->encrypteddata
		encrypted, err := utils.RsaEncrypt([]byte(yourmessage), serverPubKey, utils.RsaEncodeLimit2048)
		if err != nil {
			return ""
		}

		return base64.StdEncoding.EncodeToString(encrypted)
	}()
	ud.Signature = func() string{
		// 用自己的pri签名encrypteddata ->signature
		var hashData []byte
		hs := sha512.New()
		hs.Write([]byte(ud.Message))
		hashData = sha512.New().Sum(nil)

		var signData []byte
		signData, err = utils.RsaSign(crypto.SHA512, hashData, priKey)
		if err != nil {
			fmt.Println(err)
			return ""
		}

		return base64.StdEncoding.EncodeToString(signData)
	}()

	// 封装信息
	dispatchData := data.ServiceCenterDispatchData{}
	dispatchData.Version = "v1"
	dispatchData.Srv = "arith"
	dispatchData.Function = "add"
	dispatchData.Argv = ud

	b,err := json.Marshal(dispatchData);
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return;
	}
	fmt.Println("msg:", string(b[:]))

	ackData := data.ServiceCenterDispatchAckData{}
	for ; ;  {
		fmt.Println("Please input command: ")
		var input string
		fmt.Scanln(&input)

		fmt.Println("Execute input command: ")
		count = 0
		right = 0
		timeBegin = time.Now();
		fmt.Println("开始时间：", timeBegin)

		if input == "quit" {
			fmt.Println("I do quit")
			break;
		}else if input == "d1" {
			nethelper.CallJRPCToHttpServer("127.0.0.1:8080", "/wallet", data.MethodServiceCenterDispatch, dispatchData, &ackData)
			fmt.Println("ack==", ackData)

			bmessage, err := base64.StdEncoding.DecodeString(ackData.Value.Message)
			if err != nil {
				fmt.Println("#Error AuthData--", err.Error())
				continue
			}

			bsignature, err := base64.StdEncoding.DecodeString(ackData.Value.Signature)
			if err != nil {
				fmt.Println("#Error AuthData--", err.Error())
				continue
			}

			// 解密数据
			d2, err := utils.RsaDecrypt(bmessage, priKey, utils.RsaDecodeLimit2048)
			fmt.Println(string(d2))

			// 验证签名
			var hashData []byte
			hs := sha512.New()
			hs.Write(bmessage)
			hashData = sha512.New().Sum(nil)

			err = utils.RsaVerify(crypto.SHA512, hashData, bsignature, serverPubKey)
			fmt.Println("验证签名--", err)

		}else if input == "d2" {
			nethelper.CallJRPCToTcpServer("127.0.0.1:8090", data.MethodServiceNodeCall, dispatchData, &ackData)
			fmt.Println("ack==", ackData)
		}else if input == "d3" {
			for i := 0; i < times; i++ {
				go DoTestTcp(dispatchData, &count, &right, times)
			}
		} else if input == "d33" {

			client, err := rpc.Dial("tcp", "127.0.0.1:8090")
			if err != nil {
				log.Println("Error: ", err.Error())
				continue
			}

			for i := 0; i < times*times*2; i++ {
				go DoTestTcp2(client, dispatchData, &count, &right, times*times*2)
			}
		}else if input == "d4" {
			for i := 0; i < times; i++ {
				go DoTest(dispatchData, &count, &right, times)
			}
		} else if input == "d44" {

			addr := "127.0.0.1:8080"
			log.Println("Call JRPC to Http server...", addr)

			client, err := rpc.DialHTTPPath("tcp", addr, "/wallet")
			if err != nil {
				log.Println("Error: ", err.Error())
				continue
			}
			for i := 0; i < times*times*2; i++ {
				go DoTest2(client, dispatchData, &count, &right, times*times*2)
			}
		}
	}
}
