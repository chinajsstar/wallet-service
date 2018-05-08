package main

import (
	"fmt"
	"strings"
	"encoding/json"
	"./utils"
	"./gateway"
	"strconv"
)

func ScanLine() string {
	var c byte
	var err error
	var b []byte
	for ; err == nil; {
		_, err = fmt.Scanf("%c", &c)
		if c != '\n' && c!= '\r' {
			b = append(b, c)
		} else {
			break
		}
	}
	return string(b)
}

// 账号注册-输入--register
type ReqUserRegister struct{
	UserClass 		int `json:"user_class" comment:"用户类型，0:普通用户 1:热钱包; 2:管理员"`
	Level 			int `json:"level" comment:"级别，0：用户，100：普通管理员"`
}
// 账号注册-输出
type AckUserRegister struct{
	UserKey 		string  `json:"user_key" comment:"用户唯一标示"`
}

// 修改公钥和回调地址-输入--update profile
type ReqUserUpdateProfile struct{
	UserKey			string `json:"user_key" comment:"用户唯一标示"`
	PublicKey		string `json:"public_key" comment:"用户公钥"`
	SourceIP		string `json:"source_ip" comment:"用户源IP"`
	CallbackUrl		string `json:"callback_url" comment:"用户回调"`
}
// 修改公钥和回调地址-输出
type AckUserUpdateProfile struct{
	Status 			string `json:"status" comment:"状态"`
}

// 接口地址
const(
	httpaddrGateway = "http://127.0.0.1:8082"
)

func main() {
	var err error

	gateway.LoadRsaKeys()

	fmt.Println(">q: quit")
	fmt.Println(">rsagen name: create a rsa key")

	fmt.Println(">register userclass(0:普通用户 1:热钱包; 2:管理员) level0：用户，100：普通管理员")
	fmt.Println(">updateprofile userkey sourceip callbackurl")
	for ; ;  {
		fmt.Println("Please input command: ")
		var input string
		input = ScanLine()
		argv := strings.Split(input, " ")

		if argv[0] == "q" {
			fmt.Println("I do quit")
			break;
		}else if argv[0] == "rsagen"{
			name := ""
			if len(argv) != 2{
				fmt.Println("格式：rsagen name")
				continue
			}
			name = argv[1]

			pri := fmt.Sprintf("./private_%s.pem", name)
			pub := fmt.Sprintf("./public_%s.pem", name)
			err = utils.RsaGen(2048, pri, pub)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("rsagen ok, name-", name)
		} else if argv[0] == "register"{
			if len(argv) != 3{
				fmt.Println("格式：register userclass level")
				continue
			}

			data := ReqUserRegister{}
			v1, _ := strconv.Atoi(argv[1])
			data.UserClass = v1
			v2, _ := strconv.Atoi(argv[1])
			data.Level = v2

			b, _ := json.Marshal(data)

			d1, d2, err := gateway.SendDataWithCrypto(httpaddrGateway, string(b), "v1", "account", "register")
			fmt.Println("err==", err)
			fmt.Println("ack==", d1)
			fmt.Println("data==", string(d2))
		} else if argv[0] == "updateprofile"{
			if len(argv) != 5{
				fmt.Println("格式：updateprofile userkey pubkey sourceip callbackurl")
				continue
			}

			data := ReqUserUpdateProfile{}
			data.UserKey = argv[1]
			data.PublicKey = argv[2]
			data.SourceIP = argv[3]
			data.CallbackUrl = argv[4]

			b, _ := json.Marshal(data)

			d1, d2, err := gateway.SendDataWithCrypto(httpaddrGateway, string(b), "v1", "account", "updateprofile")
			fmt.Println("err==", err)
			fmt.Println("ack==", d1)
			fmt.Println("data==", string(d2))

		}
	}
}