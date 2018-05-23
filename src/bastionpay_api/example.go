package main

import (
	"fmt"
	"strings"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"bastionpay_api/api"
	"bastionpay_api/utils"
	"bastionpay_api/gateway"
)

func usage()  {
	fmt.Println(">q")
	fmt.Println("		quit")
	fmt.Println(">rsagen name")
	fmt.Println("		create a rsa key pair")
	fmt.Println(">setport port")
	fmt.Println("		set http listen port")
	fmt.Println(">switch cfgname")
	fmt.Println("		switch a user cfg file")
	fmt.Println(">api path jsonmessage")
	fmt.Println("		call api with json message")
	fmt.Println(">help")
	fmt.Println("		print this")
}

func main()  {
	var err error
	runDir, _ := utils.GetRunDir()

	fmt.Println("======================")
	fmt.Println("version 1.0")
	fmt.Println("======================")
	fmt.Println("执行目录：", runDir)

	usage()

	for ; ;  {
		fmt.Println("Please input command: ")
		var input string
		input = utils.ScanLine()
		argv := strings.Split(input, " ")

		if argv[0] == "q" {
			break;
		}else if argv[0] == "rsagen"{
			name := ""
			if len(argv) != 2{
				fmt.Println("格式：rsagen name")
				continue
			}
			name = argv[1]

			pub := fmt.Sprintf("%s/public_%s.pem", runDir, name)
			priv := fmt.Sprintf("%s/private_%s.pem", runDir, name)
			err = utils.RsaGen(2048, priv, pub)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("==rsagen ok==")
			fmt.Println("pubkey path: ", pub)
			fmt.Println("privkey path: ", priv)
		} else if argv[0] == "setport"{
			if len(argv) != 2{
				fmt.Println("格式：setport port")
				continue
			}
			port := argv[1]
			startHttpServer(port)

			fmt.Println("==set http port ok==")
		} else if argv[0] == "switch"{
			if len(argv) != 2{
				fmt.Println("格式：switch cfgname")
				continue
			}
			cfgName := argv[1]

			err = gateway.Init(runDir, cfgName)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("==switch cfg ok==")
		} else if argv[0] == "api"{
			if len(argv) < 2 {
				fmt.Println("格式：api path message")
				continue
			}
			path := argv[1]
			message := ""
			if len(argv) > 2 {
				message = argv[2]
			}

			fmt.Println(message)

			var ack []byte
			err := gateway.RunApi(path, []byte(message), &ack)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("err==", err)
			fmt.Println("ack==", string(ack))
		} else {
			usage()
		}
	}
}

// 处理推送
func handlePush(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println("#read body: %s", err.Error())
		return
	}

	ackData := api.UserResponseData{}
	err = json.Unmarshal(b, &ackData)
	if err != nil {
		fmt.Println("#unmarshal: %s", err.Error())
		return
	}
	fmt.Println(ackData)

	resmessage, err := gateway.Decryption(&ackData.Value)
	if err != nil {
		fmt.Println("#Decryption: %s", err.Error())
		return
	}

	fmt.Println(string(resmessage))
	return
}

// start http server
func startHttpServer(port string) error {
	fmt.Println("Start http server on: ", port)

	http.Handle("/walletcb", http.HandlerFunc(handlePush))
	go func() {
		fmt.Println("Http server routine running... ")
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			fmt.Println("#Error:", err)
			return
		}
	}()

	return nil
}