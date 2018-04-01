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
	"../base/utils"
	"crypto/sha512"
	"crypto"
	"encoding/base64"
	"strings"
	"github.com/satori/go.uuid"
	"../account_srv/user"
	"../account_srv/install"
	"errors"
	"strconv"
)

const httpaddr = "127.0.0.1:8080"
const httpaddr2 = "http://127.0.0.1:8070"

var g_admin_prikey []byte
var g_admin_pubkey []byte
var g_admin_licensekey string

var g_server_pubkey []byte

func loadRsaKeys() error {
	var err error
	g_admin_prikey, err = ioutil.ReadFile("/Users/henly.liu/workspace/private_admin.pem")
	if err != nil {
		return err
	}

	g_admin_pubkey, err = ioutil.ReadFile("/Users/henly.liu/workspace/public_admin.pem")
	if err != nil {
		return err
	}

	g_server_pubkey, err = ioutil.ReadFile("/Users/henly.liu/workspace/public_wallet.pem")
	if err != nil {
		return err
	}

	g_admin_licensekey = "cd3616ef-1ce0-47d6-81e5-832904318c90"

	return nil
}

const(
	testMessage = "{\"a\":1, \"b\":1}"
	testVersion = "v1"
	testSrv = "arith"
	testFunction = "add"
)

func sendData(message, version, srv, function string) (*data.UserResponseData, []byte, error) {
	// 用户数据
	var ud data.UserData
	ud.LicenseKey = g_admin_licensekey

	bencrypted, err := func() ([]byte, error) {
		// 用我们的pub加密message ->encrypteddata
		bencrypted, err := utils.RsaEncrypt([]byte(message), g_server_pubkey, utils.RsaEncodeLimit2048)
		if err != nil {
			return nil, err
		}
		return bencrypted, nil
	}()
	if err != nil {
		return nil, nil, err
	}

	ud.Message = base64.StdEncoding.EncodeToString(bencrypted)

	bsignature, err := func() ([]byte, error){
		// 用自己的pri签名encrypteddata ->signature
		var hashData []byte
		hs := sha512.New()
		hs.Write(bencrypted)
		hashData = hs.Sum(nil)

		bsignature, err := utils.RsaSign(crypto.SHA512, hashData, g_admin_prikey)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		return bsignature, nil
	}()
	if err != nil {
		return nil, nil, err
	}

	ud.Signature = base64.StdEncoding.EncodeToString(bsignature)

	// 封装信息
	dispatchData := data.UserRequestData{}
	dispatchData.Version = version
	dispatchData.Srv = srv
	dispatchData.Function = function
	dispatchData.Argv = ud

	////////////////////////////////////////////
	//fmt.Println("ok send msg:", dispatchData)
	ackData := &data.UserResponseData{}
	nethelper.CallJRPCToHttpServer(httpaddr, "/wallet", data.MethodCenterDispatch, dispatchData, ackData)
	//fmt.Println("ok get ack:", ackData)

	if ackData.Err != data.NoErr {
		return ackData, nil, errors.New("# got err: " + ackData.ErrMsg)
	}

	// base64 decode
	bencrypted2, err := base64.StdEncoding.DecodeString(ackData.Value.Message)
	if err != nil {
		return ackData, nil, err
	}

	bsignature2, err := base64.StdEncoding.DecodeString(ackData.Value.Signature)
	if err != nil {
		return ackData, nil, err
	}

	// 验证签名
	var hashData []byte
	hs := sha512.New()
	hs.Write([]byte(bencrypted2))
	hashData = hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, bsignature2, g_server_pubkey)
	if err != nil {
		return ackData, nil, err
	}

	// 解密数据
	d2, err := utils.RsaDecrypt(bencrypted2, g_admin_prikey, utils.RsaDecodeLimit2048)
	if err != nil {
		return ackData, nil, err
	}

	return ackData, d2, nil
}

func sendData2(message, version, srv, function string) (*data.UserResponseData, []byte, error) {
	// 用户数据
	var ud data.UserData
	ud.LicenseKey = g_admin_licensekey

	bencrypted, err := func() ([]byte, error) {
		// 用我们的pub加密message ->encrypteddata
		bencrypted, err := utils.RsaEncrypt([]byte(message), g_server_pubkey, utils.RsaEncodeLimit2048)
		if err != nil {
			return nil, err
		}
		return bencrypted, nil
	}()
	if err != nil {
		return nil, nil, err
	}

	ud.Message = base64.StdEncoding.EncodeToString(bencrypted)

	bsignature, err := func() ([]byte, error){
		// 用自己的pri签名encrypteddata ->signature
		var hashData []byte
		hs := sha512.New()
		hs.Write(bencrypted)
		hashData = hs.Sum(nil)

		bsignature, err := utils.RsaSign(crypto.SHA512, hashData, g_admin_prikey)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		return bsignature, nil
	}()
	if err != nil {
		return nil, nil, err
	}

	ud.Signature = base64.StdEncoding.EncodeToString(bsignature)

	path := "/restful"
	path += "/"+version
	path += "/"+srv
	path += "/"+function

	b, err := json.Marshal(ud)
	if err != nil {
		return nil, nil, err
	}

	body := string(b)

	////////////////////////////////////////////
	//fmt.Println("ok send msg:", dispatchData)
	ackData := &data.UserResponseData{}

	var res string
	nethelper.CallToHttpServer(httpaddr2, path, body, &res)
	//fmt.Println("ok get ack:", ackData)

	err = json.Unmarshal([]byte(res), &ackData)
	if err != nil {
		return nil, nil, err
	}

	if ackData.Err != data.NoErr {
		return ackData, nil, errors.New("# got err: " + ackData.ErrMsg)
	}

	// base64 decode
	bencrypted2, err := base64.StdEncoding.DecodeString(ackData.Value.Message)
	if err != nil {
		return ackData, nil, err
	}

	bsignature2, err := base64.StdEncoding.DecodeString(ackData.Value.Signature)
	if err != nil {
		return ackData, nil, err
	}

	// 验证签名
	var hashData []byte
	hs := sha512.New()
	hs.Write([]byte(bencrypted2))
	hashData = hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, bsignature2, g_server_pubkey)
	if err != nil {
		return ackData, nil, err
	}

	// 解密数据
	d2, err := utils.RsaDecrypt(bencrypted2, g_admin_prikey, utils.RsaDecodeLimit2048)
	if err != nil {
		return ackData, nil, err
	}

	return ackData, d2, nil
}

var timeBegin,timeEnd time.Time

func DoTest(count *int64, right *int64, times int64){
	_, d, err := sendData(testMessage, testVersion, testSrv, testFunction)

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

func DoTest2(client *rpc.Client, count *int64, right *int64, times int64){
	_, d, err := sendData(testMessage, testVersion, testSrv, testFunction)

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
	ackData := data.UserResponseData{}

	err := nethelper.CallJRPCToTcpServer("127.0.0.1:8090", data.MethodNodeCall, params, &ackData)

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
	ackData := data.UserResponseData{}
	err := nethelper.CallJRPCToTcpServerOnClient(client, data.MethodNodeCall, params, &ackData)

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
	loadRsaKeys()

	fmt.Println("Please first to login: ")
	err := func() error {
		m, err := install.LoginUser()
		if err != nil {
			fmt.Println(err)
			return err
		}

		d, err := json.Marshal(m)
		if err != nil {
			fmt.Println(err)
			return err
		}

		_, d2, err := sendData(string(d), "v1", "account", "login")
		if err != nil {
			fmt.Println(err)
			return err
		}

		uca := user.AckUserLogin{}
		err = json.Unmarshal(d2, &uca)
		if err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Println("login user ack: ", uca)

		return nil
	}()
	if err != nil {
		fmt.Println("#err:", err)
		//return
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


		fmt.Println("Execute input command: ")
		count = 0
		right = 0
		timeBegin = time.Now();
		fmt.Println("开始时间：", timeBegin)

		if argv[0] == "quit" {
			fmt.Println("I do quit")
			break;
		}else if argv[0] == "d1" {
			_, d, err := sendData(testMessage, testVersion, testSrv, testFunction)
			fmt.Println("err==", err)
			fmt.Println("ack==", string(d))

		}else if argv[0] == "d11" {
			res, d, err := sendData2(testMessage, testVersion, testSrv, testFunction)
			fmt.Println("err==", err)
			fmt.Println("res==", *res)
			fmt.Println("data==", string(d))

		}else if argv[0] == "d2" {

			var ud data.UserData
			ud.Message = testMessage

			dispatchData := data.UserRequestData{}
			dispatchData.Version = testVersion
			dispatchData.Srv = testSrv
			dispatchData.Function = testFunction
			dispatchData.Argv = ud

			ackData := data.UserResponseData{}
			nethelper.CallJRPCToTcpServer("127.0.0.1:8090", data.MethodNodeCall, dispatchData, &ackData)
			fmt.Println("ack==", ackData)
		}else if argv[0] == "d3" {

			runcounts = 100
			if len(argv) > 1 {
				runcounts, _ = strconv.Atoi(argv[1])
			}

			var ud data.UserData
			ud.Message = testMessage

			dispatchData := data.UserRequestData{}
			dispatchData.Version = testVersion
			dispatchData.Srv = testSrv
			dispatchData.Function = testFunction
			dispatchData.Argv = ud
			for i := 0; i < runcounts; i++ {
				go DoTestTcp(dispatchData, &count, &right, int64(runcounts))
			}
		} else if argv[0] == "d33" {

			runcounts = 100
			if len(argv) > 1 {
				runcounts, _ = strconv.Atoi(argv[1])
			}

			client, err := rpc.Dial("tcp", "127.0.0.1:8090")
			if err != nil {
				log.Println("Error: ", err.Error())
				continue
			}

			var ud data.UserData
			ud.Message = testMessage

			dispatchData := data.UserRequestData{}
			dispatchData.Version = testVersion
			dispatchData.Srv = testSrv
			dispatchData.Function = testFunction
			dispatchData.Argv = ud
			for i := 0; i < runcounts; i++ {
				go DoTestTcp2(client, dispatchData, &count, &right, int64(runcounts))
			}
		}else if argv[0] == "d4" {
			runcounts = 100
			if len(argv) > 1 {
				runcounts, _ = strconv.Atoi(argv[1])
			}

			for i := 0; i < runcounts; i++ {
				go DoTest(&count, &right, int64(runcounts))
			}
		} else if argv[0] == "d44" {
			runcounts = 100
			if len(argv) > 1 {
				runcounts, _ = strconv.Atoi(argv[1])
			}

			addr := "127.0.0.1:8080"
			log.Println("Call JRPC to Http server...", addr)

			client, err := rpc.DialHTTPPath("tcp", addr, "/wallet")
			if err != nil {
				log.Println("Error: ", err.Error())
				continue
			}
			for i := 0; i < runcounts; i++ {
				go DoTest2(client, &count, &right, int64(runcounts))
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
		}else if argv[0] == "test_adduser"{
			m, err := install.AddUser()
			if err != nil {
				fmt.Println(err)
				continue
			}

			d, err := json.Marshal(m)
			if err != nil {
				fmt.Println(err)
				continue
			}

			_, d2, err := sendData(string(d), "v1", "account", "create")
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(d2))

			uca := user.AckUserCreate{}
			err = json.Unmarshal(d2, &uca)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("create user ack: ", uca)
		}else if argv[0] == "test_loginuser"{
			m, err := install.LoginUser()
			if err != nil {
				fmt.Println(err)
				continue
			}

			d, err := json.Marshal(m)
			if err != nil {
				fmt.Println(err)
				continue
			}

			_, d2, err := sendData(string(d), "v1", "account", "login")
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(d2))

			uca := user.AckUserLogin{}
			err = json.Unmarshal(d2, &uca)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("login user ack: ", uca)
		}else if argv[0] == "test_listusers"{
			m, err := install.ListUsers()
			if err != nil {
				fmt.Println(err)
				continue
			}

			d, err := json.Marshal(m)
			if err != nil {
				fmt.Println(err)
				continue
			}

			_, d2, err := sendData(string(d), "v1", "account", "listusers")
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(d2))

			uca := user.AckUserList{}
			err = json.Unmarshal(d2, &uca)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("list users ack: ", uca)
		}
	}
}
