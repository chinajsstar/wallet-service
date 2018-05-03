package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"log"
	"io/ioutil"
	"api_router/base/utils"
	"crypto/sha512"
	"crypto"
	"encoding/base64"
	"strings"
	"errors"
	"bytes"
)

/////////////////////////////////////////////////////
// 网关通用结构
// input/output data/value
// when input data, user encode and sign data, server decode and verify;
// when output value, server encode and sign data, user decode and verify;
type userData struct {
	// user unique key
	UserKey string `json:"user_key"`
	// message = origin data -> rsa encode -> base64
	Message    string `json:"message"`
	// signature = origin data -> sha512 -> rsa sign -> base64
	Signature  string `json:"signature"`
}

// input/output method
type userMethod struct {
	Version     string `json:"version"`   // srv version
	Srv     	string `json:"srv"`	  	  // srv name
	Function  	string `json:"function"`  // srv function
}

// user response/push data
type userResponseData struct{
	Method		userMethod 	`json:"method"`	// response method
	Err     	int    		`json:"err"`    // error code
	ErrMsg  	string 		`json:"errmsg"` // error message
	Value   	userData 	`json:"value"` 	// response data
}

var (
	// 接口地址
	httpaddrGateway = "http://127.0.0.1:8082"

	// 客户私钥
	user_prikey []byte

	// 客户公钥
	user_pubkey []byte

	// 客户user key
	user_key string

	// 服务公钥
	server_pubkey []byte
)

// 加载数据
func loadRsaKeys(curDir, name, userKey string) error {
	var err error
	user_prikey, err = ioutil.ReadFile(curDir + "/private_" + name + ".pem")
	if err != nil {
		return err
	}

	user_pubkey, err = ioutil.ReadFile(curDir + "/public_" + name + ".pem")
	if err != nil {
		return err
	}

	server_pubkey, err = ioutil.ReadFile(curDir + "/public.pem")
	if err != nil {
		return err
	}

	user_key = userKey
	return nil
}

// 发送http请求
func callToHttpServer(addr string, path string, body string, res *string) error {
	url := addr + path
	contentType := "application/json;charset=utf-8"

	b := []byte(body)
	b2 := bytes.NewBuffer(b)

	resp, err := http.Post(url, contentType, b2)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	*res = string(content)
	return nil
}

// 加密签名数据
func encryptData(message string) (*userData, error) {
	// 用户数据
	ud := &userData{}
	ud.UserKey = user_key

	bencrypted, err := func() ([]byte, error) {
		bencrypted, err := utils.RsaEncrypt([]byte(message), server_pubkey, utils.RsaEncodeLimit2048)
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
		var hashData []byte
		hs := sha512.New()
		hs.Write(bencrypted)
		hashData = hs.Sum(nil)

		bsignature, err := utils.RsaSign(crypto.SHA512, hashData, user_prikey)
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

	return ud, nil
}

// 验证解密数据
func decryptData(ud *userData) (string, error) {
	var d2 []byte
	// base64 decode
	bencrypted2, err := base64.StdEncoding.DecodeString(ud.Message)
	if err != nil {
		return "", err
	}

	bsignature2, err := base64.StdEncoding.DecodeString(ud.Signature)
	if err != nil {
		return "", err
	}

	// 验证签名
	var hashData []byte
	hs := sha512.New()
	hs.Write([]byte(bencrypted2))
	hashData = hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, bsignature2, server_pubkey)
	if err != nil {
		return "", err
	}

	// 解密数据
	d2, err = utils.RsaDecrypt(bencrypted2, user_prikey, utils.RsaDecodeLimit2048)
	if err != nil {
		return "", err
	}

	return string(d2), nil
}

// 请求使用加密签名
func sendDataWithCrypto(addr, message, version, srv, function string) (*userResponseData, []byte, error) {
	ud, err := encryptData(message)
	if err != nil {
		log.Println("#encrypt: %s", err.Error())
		return nil, nil, err
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
	ackData := &userResponseData{}

	var res string
	callToHttpServer(addr, path, body, &res)
	fmt.Println("ok get ack:", res)

	err = json.Unmarshal([]byte(res), &ackData)
	if err != nil {
		return nil, nil, err
	}

	if ackData.Err != 0 {
		fmt.Println("err: ", ackData.Err, "-msg: ", ackData.ErrMsg)
		return ackData, nil, errors.New("# got err: " + ackData.ErrMsg)
	}

	resmessage, err := decryptData(&ackData.Value)
	if err != nil {
		return nil, nil, err
	}

	return ackData, []byte(resmessage), nil
}

// 请求不使用加密签名
func sendDataWithNoCrypto(addr, message, version, srv, function string) (*userResponseData, []byte, error) {
	// 用户数据
	ud := userData{}
	ud.UserKey = user_key
	ud.Message = message

	path := "/wallettest"
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
	ackData := &userResponseData{}

	var res string
	callToHttpServer(addr, path, body, &res)
	fmt.Println("ok get ack:", res)

	err = json.Unmarshal([]byte(res), &ackData)
	if err != nil {
		return nil, nil, err
	}

	if ackData.Err != 0 {
		fmt.Println("err: ", ackData.Err, "-msg: ", ackData.ErrMsg)
		return ackData, nil, errors.New("# got err: " + ackData.ErrMsg)
	}

	uda := ackData.Value

	return ackData, []byte(uda.Message), nil
}

// http
// 不加密示例：
// curl -d '{"id":-1}' http://localhost:8077/wallettest/v1/account/listusers
// curl -d '{"a":2, "b":1}' http://localhost:8077/wallettest/v1/arith/add

// 加密示例：message为加密数据
// curl -d '{"license_key":"719101fe-93a0-44e5-909b-84a6e7fcb132", "signature":"", "message":"{\"id\":-1}"' http://localhost:8077/wallet/v1/account/listusers
// curl -d '{"license_key":"719101fe-93a0-44e5-909b-84a6e7fcb132", "signature":"", "message":"{\"a\":2, \"b\":1}"' http://localhost:8077/wallet/v1/arith/add
func main() {
	var err error

	// 目录
	curDir, _ := utils.GetCurrentDir()
	log.Println("当前目录：%s", curDir)

	fmt.Println(">q: quit")
	fmt.Println(">rsagen name: create a rsa key")
	fmt.Println(">set addr: set remote http addr")
	fmt.Println(">load port name userkey: load")

	fmt.Println(">api srv function message")
	fmt.Println(">testapi srv function message")
	for ; ;  {
		fmt.Println("Please input command: ")
		var input string
		input = utils.ScanLine()
		argv := strings.Split(input, " ")

		if argv[0] == "q" {
			fmt.Println("I do quit")
			break;
		}else if argv[0] == "rsagen"{
			user := ""
			if len(argv) != 2{
				fmt.Println("格式：rsagen name")
				continue
			}
			user = argv[1]

			pri := fmt.Sprintf("%s/private_%s.pem", curDir, user)
			pub := fmt.Sprintf("%s/public_%s.pem", curDir, user)
			err = utils.RsaGen(2048, pri, pub)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("rsagen ok, user-", user)
		} else if argv[0] == "set"{
			var addr string
			if len(argv) != 2{
				fmt.Println("格式：set addr")
				continue
			}
			addr = argv[1]
			httpaddrGateway = addr
		} else if argv[0] == "load"{
			var port, name, userkey string
			if len(argv) != 4{
				fmt.Println("格式：load port name userkey")
				continue
			}
			port = argv[1]
			name = argv[2]
			userkey = argv[3]

			// 加载数据
			err := loadRsaKeys(curDir, name, userkey)
			if err != nil {
				log.Fatalf("%s", err.Error())
				return
			}

			// 启动回调地址
			startHttpServer(port)
		} else if argv[0] == "api"{
			var srv, function, message string
			if len(argv) != 4{
				fmt.Println("格式：api srv function message")
				continue
			}
			srv = argv[1]
			function = argv[2]
			message = argv[3]

			go func() {
				_, d, err := sendDataWithCrypto(httpaddrGateway, message, "v1", srv, function)
				fmt.Println("err==", err)
				fmt.Println("ack==", string(d))
			}()
		}else if argv[0] == "testapi"{
			var srv, function, message string
			if len(argv) != 4{
				fmt.Println("格式：testapi srv function message")
				continue
			}
			srv = argv[1]
			function = argv[2]
			message = argv[3]

			go func() {
				_, d, err := sendDataWithNoCrypto(httpaddrGateway, message, "v1", srv, function)
				fmt.Println("err==", err)
				fmt.Println("ack==", string(d))
			}()
		}
	}
}

// 处理推送
func handlePush(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("#read body: %s", err.Error())
		return
	}
	//fmt.Println(string(b))

	ackData := userResponseData{}
	err = json.Unmarshal(b, &ackData.Value)
	if err != nil {
		log.Println("#unmarshal: %s", err.Error())
		return
	}

	resmessage, err := decryptData(&ackData.Value)
	if err != nil {
		log.Println("#decryptData2: %s", err.Error())
		return
	}

	fmt.Println(resmessage)
	return
}

// start http server
func startHttpServer(port string) error {
	log.Println("Start http server on ", port)

	http.Handle("/walletcb", http.HandlerFunc(handlePush))
	go func() {
		log.Println("Http server routine running... ")
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			fmt.Println("#Error:", err)
			return
		}
	}()

	return nil
}