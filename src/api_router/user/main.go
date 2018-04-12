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
	"../account_srv/user"
	"../account_srv/install"
	"errors"
	"strconv"
	"golang.org/x/net/websocket"
	l4g "github.com/alecthomas/log4go"
)

const wsaddrGateway = "ws://127.0.0.1:8088/ws"
const tcpaddrGateway = "127.0.0.1:8081"
const httpaddrGateway = "http://127.0.0.1:8080"
const httpaddrNigix = "http://127.0.0.1:8070"

var G_admin_prikey []byte
var G_admin_pubkey []byte
var G_admin_licensekey string

var G_henly_prikey []byte
var G_henly_pubkey []byte
var G_henly_licensekey string

var G_server_pubkey []byte

var wsconn *websocket.Conn

func LoadRsaKeys() error {
	var err error
	G_admin_prikey, err = ioutil.ReadFile("/Users/henly.liu/workspace/private_admin.pem")
	if err != nil {
		return err
	}

	G_admin_pubkey, err = ioutil.ReadFile("/Users/henly.liu/workspace/public_admin.pem")
	if err != nil {
		return err
	}

	G_henly_prikey, err = ioutil.ReadFile("/Users/henly.liu/workspace/private_henly.pem")
	if err != nil {
		return err
	}

	G_henly_pubkey, err = ioutil.ReadFile("/Users/henly.liu/workspace/public_henly.pem")
	if err != nil {
		return err
	}

	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	accountDir := appDir + "/account"
	G_server_pubkey, err = ioutil.ReadFile(accountDir + "/public.pem")
	if err != nil {
		return err
	}

	G_admin_licensekey = "4a871a62-1924-4a1b-b63a-d244774747e1"

	G_henly_licensekey = "524faf3a-b6a0-42ce-9c49-9c07b66aa835"

	return nil
}

const(
	testMessage = "{\"a\":1, \"b\":1}"
	testVersion = "v1"
	testSrv = "arith"
	testFunction = "add"
	dev = false
)

func sendData2(addr, message, version, srv, function string) (*data.UserResponseData, []byte, error) {
	// 用户数据
	var ud data.UserData
	ud.LicenseKey = G_admin_licensekey

	if dev == false{
		bencrypted, err := func() ([]byte, error) {
			// 用我们的pub加密message ->encrypteddata
			bencrypted, err := utils.RsaEncrypt([]byte(message), G_server_pubkey, utils.RsaEncodeLimit2048)
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

			bsignature, err := utils.RsaSign(crypto.SHA512, hashData, G_admin_prikey)
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
	}else{
		ud.Message = message
	}

	path := "/wallet"
	path += "/"+version
	path += "/"+srv
	path += "/"+function

	b, err := json.Marshal(ud)
	if err != nil {
		return nil, nil, err
	}

	body := string(b)

	////////////////////////////////////////////
	fmt.Println("ok send msg:", body)
	ackData := &data.UserResponseData{}

	var res string
	nethelper.CallToHttpServer(addr, path, body, &res)
	//fmt.Println("ok get ack:", res)

	err = json.Unmarshal([]byte(res), &ackData)
	if err != nil {
		return nil, nil, err
	}

	if ackData.Err != data.NoErr {
		fmt.Println("err: ", ackData.Err, "-msg: ", ackData.ErrMsg)
		return ackData, nil, errors.New("# got err: " + ackData.ErrMsg)
	}

	var d2 []byte
	if dev == false {
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

		err = utils.RsaVerify(crypto.SHA512, hashData, bsignature2, G_server_pubkey)
		if err != nil {
			return ackData, nil, err
		}

		// 解密数据
		d2, err = utils.RsaDecrypt(bencrypted2, G_admin_prikey, utils.RsaDecodeLimit2048)
		if err != nil {
			return ackData, nil, err
		}
	}else{
		d2 = []byte(ackData.Value.Message)
	}

	return ackData, d2, nil
}

var timeBegin,timeEnd time.Time

func DoTest(count *int64, right *int64, times int64){
	_, d, err := sendData2(httpaddrGateway, testMessage, testVersion, testSrv, testFunction)

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

func DoTest2(count *int64, right *int64, times int64){
	_, d, err := sendData2(httpaddrGateway, testMessage, testVersion, testSrv, testFunction)

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

	err := nethelper.CallJRPCToTcpServer(tcpaddrGateway, data.MethodCenterDispatch, params, &ackData)

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
	err := nethelper.CallJRPCToTcpServerOnClient(client, data.MethodCenterDispatch, params, &ackData)

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

// http
// curl -d '{"license_key":"719101fe-93a0-44e5-909b-84a6e7fcb132", "signature":"", "message":"{\"a\":2, \"b\":1}"}' http://localhost:8080/wallet/v1/arith/add
// curl -d '{"a":2, "b":1}' http://localhost:8077/wallet/v1/arith/add
// curl -d '{"user_name":"henly", "password":"123456"}' http://localhost:8077/wallet/v1/account/login
// curl -d '{"id":-1}' http://localhost:8077/wallet/v1/account/listusers
func main() {
	appDir, _:= utils.GetAppDir()
	appDir += "/SuperWallet"

	//l4g.LoadConfiguration(appDir + "/log.xml")
	l4g.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter())
	l4g.Info("L4g Begin, the time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	defer l4g.Info("L4g End, the time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	defer l4g.Close()

	// 目录
	curDir, _ := utils.GetCurrentDir()
	runDir, _ := utils.GetRunDir()
	l4g.Info("当前目录：%s", curDir)
	l4g.Info("执行目录：%s", runDir)
	// 加载服务器公钥
	err := LoadRsaKeys()
	if err != nil {
		l4g.Error("%s", err.Error())
		return
	}

	fmt.Println("Please first to login: ")
	err = func() error {
		m, err := install.LoginUser()
		if err != nil {
			l4g.Error("%s", err.Error())
			return err
		}

		d, err := json.Marshal(m)
		if err != nil {
			l4g.Error("%s", err.Error())
			return err
		}

		_, d2, err := sendData2(httpaddrGateway, string(d), "v1", "account", "login")
		if err != nil {
			l4g.Error("%s", err.Error())
			return err
		}

		uca := user.AckUserLogin{}
		err = json.Unmarshal(d2, &uca)
		if err != nil {
			l4g.Error("%s", err.Error())
			return err
		}

		l4g.Info("Login ack: ", uca)

		return nil
	}()
	if err != nil {
		l4g.Error("Login err: %s", err.Error())
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

		if argv[0] == "quit" {
			fmt.Println("I do quit")
			break;
		}else if argv[0] == "d1" {
			_, d, err := sendData2(httpaddrGateway, testMessage, testVersion, testSrv, testFunction)
			fmt.Println("err==", err)
			fmt.Println("ack==", string(d))

		}else if argv[0] == "d11" {
			res, d, err := sendData2(httpaddrNigix, testMessage, testVersion, testSrv, testFunction)
			fmt.Println("err==", err)
			fmt.Println("res==", *res)
			fmt.Println("data==", string(d))

		}else if argv[0] == "d2" {

			var ud data.UserData
			ud.Message = testMessage

			dispatchData := data.UserRequestData{}
			dispatchData.Method.Version = testVersion
			dispatchData.Method.Srv = testSrv
			dispatchData.Method.Function = testFunction
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
			dispatchData.Method.Version = testVersion
			dispatchData.Method.Srv = testSrv
			dispatchData.Method.Function = testFunction
			dispatchData.Argv = ud
			for i := 0; i < runcounts; i++ {
				go DoTestTcp(&dispatchData, &count, &right, int64(runcounts))
			}
		} else if argv[0] == "d33" {

			runcounts = 100
			if len(argv) > 1 {
				runcounts, _ = strconv.Atoi(argv[1])
			}

			client, err := rpc.Dial("tcp", tcpaddrGateway)
			if err != nil {
				log.Println("Error: ", err.Error())
				continue
			}

			var ud data.UserData
			ud.Message = testMessage

			dispatchData := data.UserRequestData{}
			dispatchData.Method.Version = testVersion
			dispatchData.Method.Srv = testSrv
			dispatchData.Method.Function = testFunction
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
/*
			addr := "127.0.0.1:8080"
			log.Println("Call JRPC to Http server...", addr)

			client, err := rpc.DialHTTPPath("tcp", addr, "/wallet")
			if err != nil {
				log.Println("Error: ", err.Error())
				continue
			}*/
			for i := 0; i < runcounts; i++ {
				go DoTest2(&count, &right, int64(runcounts))
			}
		}else if argv[0] == "rsagen"{
			user := ""
			if len(argv) < 2{
				fmt.Println("输入名称")
				continue
			}
			user = argv[1]

			pri := fmt.Sprintf("/Users/henly.liu/workspace/private_%s.pem", user)
			pub := fmt.Sprintf("/Users/henly.liu/workspace/public_%s.pem", user)
			err = utils.RsaGen(2048, pri, pub)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("rsagen ok, user-", user)
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

			_, d2, err := sendData2(httpaddrGateway, string(d), "v1", "account", "create")
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

			_, d2, err := sendData2(httpaddrGateway, string(d), "v1", "account", "login")
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

			_, d2, err := sendData2(httpaddrGateway, string(d), "v1", "account", "listusers")
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
		}else if argv[0] == "ws"{
			startWsClient()
		}else if argv[0] == "wslogin"{
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

			var ud data.UserData
			ud.LicenseKey = G_henly_licensekey
			encryptRequestData(string(d), G_henly_prikey, &ud)

			dispatchData := data.UserRequestData{}
			dispatchData.Method.Version = "v1"
			dispatchData.Method.Srv = "account"
			dispatchData.Method.Function = "login"
			dispatchData.Argv = ud

			d, err = json.Marshal(dispatchData)

			if wsconn != nil {
				websocket.Message.Send(wsconn, string(d))
			}
		}else if argv[0] == "newaddress"{
			s := "{\"user_id\":\"0001\",\"method\":\"new_address\",\"params\":{\"id\":\"1\",\"symbol\":\"eth\",\"count\":10}}"

			_, d2, err := sendData2(httpaddrGateway, s, "v1", "xxxx", "new_address")
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

			fmt.Println("new address ack: ", uca)
		}
	}
}

func startWsClient() *websocket.Conn {
	conn, err := websocket.Dial(wsaddrGateway, "", "test://wallet/")
	if err != nil {
		fmt.Println("#error", err)
		return nil
	}

	go func(conn *websocket.Conn) {
		for ; ; {
			var data string
			err := websocket.Message.Receive(conn, &data)
			if err != nil {
				fmt.Println("read failed:", err)
				break
			}

			pData, err:= decryptPushData(data, G_henly_prikey)
			if pData != nil{
				fmt.Println("pData:", pData)
			}else{
				fmt.Println("read:", data)
			}
		}
	}(conn)

	wsconn = conn
	return conn
}

func encryptRequestData(message string, prikey []byte, userData *data.UserData) (error) {
	// 用户数据
	bencrypted, err := func() ([]byte, error) {
		// 用我们的pub加密message ->encrypteddata
		bencrypted, err := utils.RsaEncrypt([]byte(message), G_server_pubkey, utils.RsaEncodeLimit2048)
		if err != nil {
			return nil, err
		}
		return bencrypted, nil
	}()
	if err != nil {
		return err
	}

	userData.Message = base64.StdEncoding.EncodeToString(bencrypted)

	bsignature, err := func() ([]byte, error) {
		// 用自己的pri签名encrypteddata ->signature
		var hashData []byte
		hs := sha512.New()
		hs.Write(bencrypted)
		hashData = hs.Sum(nil)

		bsignature, err := utils.RsaSign(crypto.SHA512, hashData, prikey)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		return bsignature, nil
	}()
	if err != nil {
		return err
	}

	userData.Signature = base64.StdEncoding.EncodeToString(bsignature)

	return nil
}

func decryptResponseData(res string, prikey []byte) (*data.UserResponseData, error) {
	ackData := &data.UserResponseData{}
	err := json.Unmarshal([]byte(res), &ackData)
	if err != nil {
		return nil, err
	}

	if ackData.Err != data.NoErr {
		return ackData, errors.New("# got err: " + ackData.ErrMsg)
	}

	// base64 decode
	bencrypted2, err := base64.StdEncoding.DecodeString(ackData.Value.Message)
	if err != nil {
		return ackData, err
	}

	bsignature2, err := base64.StdEncoding.DecodeString(ackData.Value.Signature)
	if err != nil {
		return ackData, err
	}

	// 验证签名
	var hashData []byte
	hs := sha512.New()
	hs.Write([]byte(bencrypted2))
	hashData = hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, bsignature2, G_server_pubkey)
	if err != nil {
		return ackData, err
	}

	// 解密数据
	d, err := utils.RsaDecrypt(bencrypted2, prikey, utils.RsaDecodeLimit2048)
	if err != nil {
		return ackData, err
	}
	ackData.Value.Message = string(d)

	return ackData, nil
}

func decryptPushData(res string, prikey []byte) (*data.UserResponseData, error) {
	ackData := &data.UserResponseData{}
	err := json.Unmarshal([]byte(res), &ackData)
	if err != nil {
		return nil, err
	}

	// base64 decode
	bencrypted2, err := base64.StdEncoding.DecodeString(ackData.Value.Message)
	if err != nil {
		return ackData, err
	}

	bsignature2, err := base64.StdEncoding.DecodeString(ackData.Value.Signature)
	if err != nil {
		return ackData, err
	}

	// 验证签名
	var hashData []byte
	hs := sha512.New()
	hs.Write([]byte(bencrypted2))
	hashData = hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, bsignature2, G_server_pubkey)
	if err != nil {
		return ackData, err
	}

	// 解密数据
	d, err := utils.RsaDecrypt(bencrypted2, prikey, utils.RsaDecodeLimit2048)
	if err != nil {
		return ackData, err
	}
	ackData.Value.Message = string(d)

	return ackData, nil
}