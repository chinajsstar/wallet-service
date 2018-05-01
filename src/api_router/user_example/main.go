package main

import (
	"api_router/base/nethelper"
	"api_router/base/data"
	"fmt"
	"encoding/json" // for json get
	"net/http"
	"log"
	"io/ioutil"
	"api_router/base/utils"
	"crypto/sha512"
	"crypto"
	"encoding/base64"
	"strings"
	"errors"
)

var httpaddrGateway = "http://127.0.0.1:8082"

var user_prikey []byte
var user_pubkey []byte
var user_licensekey string

var G_server_pubkey []byte

func LoadRsaKeys(curDir, name, userKey string) error {
	var err error
	user_prikey, err = ioutil.ReadFile(curDir + "/private_" + name + ".pem")
	if err != nil {
		return err
	}

	user_pubkey, err = ioutil.ReadFile(curDir + "/public_" + name + ".pem")
	if err != nil {
		return err
	}

	G_server_pubkey, err = ioutil.ReadFile(curDir + "/public.pem")
	if err != nil {
		return err
	}

	//user_licensekey = "1c75c668-f1ab-474b-9dae-9ed7950604b4"
	user_licensekey = userKey
	return nil
}

func encryptData2(message string) (*data.UserData, error) {
	// 用户数据
	ud := &data.UserData{}
	ud.UserKey = user_licensekey

	bencrypted, err := func() ([]byte, error) {
		// 用我们的pub加密message ->encrypteddata
		bencrypted, err := utils.RsaEncrypt([]byte(message), G_server_pubkey, utils.RsaEncodeLimit2048)
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

func decryptData2(ud1 *data.UserData) (*data.UserData, error) {
	ud := &data.UserData{}

	var d2 []byte
	// base64 decode
	bencrypted2, err := base64.StdEncoding.DecodeString(ud1.Message)
	if err != nil {
		return nil, err
	}

	bsignature2, err := base64.StdEncoding.DecodeString(ud1.Signature)
	if err != nil {
		return nil, err
	}

	// 验证签名
	var hashData []byte
	hs := sha512.New()
	hs.Write([]byte(bencrypted2))
	hashData = hs.Sum(nil)

	err = utils.RsaVerify(crypto.SHA512, hashData, bsignature2, G_server_pubkey)
	if err != nil {
		return nil, err
	}

	// 解密数据
	d2, err = utils.RsaDecrypt(bencrypted2, user_prikey, utils.RsaDecodeLimit2048)
	if err != nil {
		return nil, err
	}

	*ud = *ud1
	ud.Message = string(d2)

	return ud, nil
}

func sendData2(addr, message, version, srv, function string) (*data.UserResponseData, []byte, error) {
	// 用户数据
	ud, err := encryptData2(message)
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
	ackData := &data.UserResponseData{}

	var res string
	nethelper.CallToHttpServer(addr, path, body, &res)
	fmt.Println("ok get ack:", res)

	err = json.Unmarshal([]byte(res), &ackData)
	if err != nil {
		return nil, nil, err
	}

	if ackData.Err != data.NoErr {
		fmt.Println("err: ", ackData.Err, "-msg: ", ackData.ErrMsg)
		return ackData, nil, errors.New("# got err: " + ackData.ErrMsg)
	}

	uda, err := decryptData2(&ackData.Value)
	if err != nil {
		return nil, nil, err
	}

	return ackData, []byte(uda.Message), nil
}

func sendData3(addr, message, version, srv, function string) (*data.UserResponseData, []byte, error) {
	// 用户数据
	ud := data.UserData{}
	ud.UserKey = user_licensekey
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
	ackData := &data.UserResponseData{}

	var res string
	nethelper.CallToHttpServer(addr, path, body, &res)
	fmt.Println("ok get ack:", res)

	err = json.Unmarshal([]byte(res), &ackData)
	if err != nil {
		return nil, nil, err
	}

	if ackData.Err != data.NoErr {
		fmt.Println("err: ", ackData.Err, "-msg: ", ackData.ErrMsg)
		return ackData, nil, errors.New("# got err: " + ackData.ErrMsg)
	}

	uda := ackData.Value

	return ackData, []byte(uda.Message), nil
}

// http
// curl -d '{"license_key":"719101fe-93a0-44e5-909b-84a6e7fcb132", "signature":"", "message":"{\"a\":2, \"b\":1}"}' http://localhost:8080/wallet/v1/arith/add
// curl -d '{"a":2, "b":1}' http://localhost:8077/wallet/v1/arith/add
// curl -d '{"user_name":"henly", "password":"123456"}' http://localhost:8077/wallet/v1/account/login
// curl -d '{"id":-1}' http://localhost:8077/wallet/v1/account/listusers
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
			err := LoadRsaKeys(curDir, name, userkey)
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
				_, d, err := sendData2(httpaddrGateway, message, "v1", srv, function)
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
				_, d, err := sendData3(httpaddrGateway, message, "v1", srv, function)
				fmt.Println("err==", err)
				fmt.Println("ack==", string(d))
			}()
		}
	}
}

// http handler
func handleCheck(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("#read body: %s", err.Error())
		return
	}
	//fmt.Println(string(b))

	ackData := data.UserResponseData{}
	err = json.Unmarshal(b, &ackData.Value)
	if err != nil {
		log.Println("#unmarshal: %s", err.Error())
		return
	}

	uda, err := decryptData2(&ackData.Value)
	if err != nil {
		log.Println("#decryptData2: %s", err.Error())
		return
	}

	fmt.Println(uda.Message)
	return
}
// start http server
func startHttpServer(port string) error {
	// http
	log.Println("Start http server on ", port)

	http.Handle("/walletcb", http.HandlerFunc(handleCheck))

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