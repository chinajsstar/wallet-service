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
	"strings"
	"github.com/satori/go.uuid"
	"../user_srv/user"
	"errors"
	"./admin"
)


func sendData(serverPubKey []byte, message, version, srv, function string) ([]byte, error) {
	priPath := fmt.Sprintf("/Users/henly.liu/workspace/private_%s.pem", "1")

	var err error
	var priKey []byte
	priKey, err = ioutil.ReadFile(priPath)
	if err != nil {
		return nil, err
	}

	// 用户数据
	var ud data.UserData
	ud.LicenseKey = "c9730876-6f26-4bcb-8e8e-38b384280644"

	bencrypted, err := func() ([]byte, error) {
		// 用我们的pub加密message ->encrypteddata
		bencrypted, err := utils.RsaEncrypt([]byte(message), serverPubKey, utils.RsaEncodeLimit2048)
		if err != nil {
			return nil, err
		}
		return bencrypted, nil
	}()
	if err != nil {
		return nil, err
	}

	ud.Message = base64.StdEncoding.EncodeToString(bencrypted)

	bsignature, err := func() ([]byte, error){
		// 用自己的pri签名encrypteddata ->signature
		var hashData []byte
		hs := sha512.New()
		hs.Write(bencrypted)
		hashData = hs.Sum(nil)

		bsignature, err := utils.RsaSign(crypto.SHA512, hashData, priKey)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		return bsignature, nil
	}()
	if err != nil {
		return nil, err
	}

	ud.Signature = base64.StdEncoding.EncodeToString(bsignature)

	// 封装信息
	dispatchData := data.ServiceCenterDispatchData{}
	dispatchData.Version = version
	dispatchData.Srv = srv
	dispatchData.Function = function
	dispatchData.Argv = ud

	////////////////////////////////////////////
	fmt.Println("ok send msg:", dispatchData)
	ackData := data.ServiceCenterDispatchAckData{}
	nethelper.CallJRPCToHttpServer("127.0.0.1:8080", "/wallet", data.MethodServiceCenterDispatch, dispatchData, &ackData)
	fmt.Println("ok get ack:", ackData)

	if ackData.Err != data.NoErr {
		return nil, errors.New("# got err: " + ackData.ErrMsg)
	}

	// base64 decode
	bencrypted2, err := base64.StdEncoding.DecodeString(ackData.Value.Message)
	if err != nil {
		return nil, err
	}

	bsignature2, err := base64.StdEncoding.DecodeString(ackData.Value.Signature)
	if err != nil {
		return nil, err
	}

	// 验证签名
	var hashData []byte
	hs := sha512.New()
	hs.Write([]byte(bencrypted2))
	hashData = hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, bsignature2, serverPubKey)
	if err != nil {
		return nil, err
	}

	// 解密数据
	d2, err := utils.RsaDecrypt(bencrypted2, priKey, utils.RsaDecodeLimit2048)
	if err != nil {
		return nil, err
	}

	return d2, nil
}

const(
	testMessage = "{\"a\":1, \"b\":1}"
	testVersion = "v1"
	testSrv = "arith"
	testFunction = "add"
)

var timeBegin,timeEnd time.Time

func DoTest(serPubKey []byte, count *int64, right *int64, times int64){
	d, err := sendData(serPubKey, testMessage, testVersion, testSrv, testFunction)

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

func DoTest2(client *rpc.Client, serPubKey []byte, count *int64, right *int64, times int64){
	d, err := sendData(serPubKey, testMessage, testVersion, testSrv, testFunction)

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
	// 加载服务器公钥
	var err error
	serverPubPath := fmt.Sprintf("/Users/henly.liu/workspace/public.pem")
	var serverPubKey []byte
	serverPubKey, err = ioutil.ReadFile(serverPubPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 测试次数
	const times = 100;
	var count, right int64
	count = 0
	right = 0

	for ; ;  {
		fmt.Println("Please input command: ")
		var input string
		input = utils.ScanLine()

		argv := strings.Split(input, " ")


		fmt.Println("Execute input command: ")
		count = 0
		right = 0
		timeBegin = time.Now();
		fmt.Println("开始时间：", timeBegin)

		if argv[0] == "quit" {
			fmt.Println("I do quit")
			break;
		}else if argv[0] == "d1" {
			d, err := sendData(serverPubKey, testMessage, testVersion, testSrv, testFunction)
			fmt.Println("err==", err)
			fmt.Println("ack==", string(d))

		}else if argv[0] == "d2" {

			var ud data.UserData
			ud.Message = testMessage

			dispatchData := data.ServiceCenterDispatchData{}
			dispatchData.Version = testVersion
			dispatchData.Srv = testSrv
			dispatchData.Function = testFunction
			dispatchData.Argv = ud

			ackData := data.ServiceCenterDispatchAckData{}
			nethelper.CallJRPCToTcpServer("127.0.0.1:8090", data.MethodServiceNodeCall, dispatchData, &ackData)
			fmt.Println("ack==", ackData)
		}else if argv[0] == "d3" {

			var ud data.UserData
			ud.Message = testMessage

			dispatchData := data.ServiceCenterDispatchData{}
			dispatchData.Version = testVersion
			dispatchData.Srv = testSrv
			dispatchData.Function = testFunction
			dispatchData.Argv = ud
			for i := 0; i < times; i++ {
				go DoTestTcp(dispatchData, &count, &right, times)
			}
		} else if argv[0] == "d33" {

			client, err := rpc.Dial("tcp", "127.0.0.1:8090")
			if err != nil {
				log.Println("Error: ", err.Error())
				continue
			}

			var ud data.UserData
			ud.Message = testMessage

			dispatchData := data.ServiceCenterDispatchData{}
			dispatchData.Version = testVersion
			dispatchData.Srv = testSrv
			dispatchData.Function = testFunction
			dispatchData.Argv = ud
			for i := 0; i < times*times*2; i++ {
				go DoTestTcp2(client, dispatchData, &count, &right, times*times*2)
			}
		}else if argv[0] == "d4" {
			for i := 0; i < times; i++ {
				go DoTest(serverPubKey, &count, &right, times)
			}
		} else if argv[0] == "d44" {

			addr := "127.0.0.1:8080"
			log.Println("Call JRPC to Http server...", addr)

			client, err := rpc.DialHTTPPath("tcp", addr, "/wallet")
			if err != nil {
				log.Println("Error: ", err.Error())
				continue
			}
			for i := 0; i < times*times*2; i++ {
				go DoTest2(client, serverPubKey, &count, &right, times*times*2)
			}
		}else if argv[0] == "rsagen"{
			u, err := uuid.NewV4()
			if err != nil {
				fmt.Println("#uuid create failed")
				continue
			}
			uuid := u.String()

			pri := fmt.Sprintf("/Users/henly.liu/workspace/private_%s.pem", uuid)
			pub := fmt.Sprintf("/Users/henly.liu/workspace/public_%s.pem", uuid)
			err = utils.RsaGen(2048, pri, pub)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("rsagen ok, uuid-", uuid)
		}else if argv[0] == "adduser"{

			m, err := admin.AddUser()
			if err != nil {
				fmt.Println(err)
				continue
			}

			d2, err := sendData(serverPubKey, m, "v1", "user", "create")
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(d2))

			uca := user.UserCreateAck{}
			err = json.Unmarshal(d2, &uca)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("create user ack: ", uca)
		}else if argv[0] == "loginuser"{

			m, err := admin.LoginUser()
			if err != nil {
				fmt.Println(err)
				continue
			}

			d2, err := sendData(serverPubKey, m, "v1", "user", "login")
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(d2))

			uca := user.UserLoginAck{}
			err = json.Unmarshal(d2, &uca)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("login user ack: ", uca)
		}else if argv[0] == "listusers"{

			m, err := admin.ListUsers()
			if err != nil {
				fmt.Println(err)
				continue
			}

			d2, err := sendData(serverPubKey, m, "v1", "user", "listusers")
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(d2))

			uca := user.UserListAck{}
			err = json.Unmarshal(d2, &uca)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("list users ack: ", uca)
		}
	}
}
