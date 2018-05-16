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
	"bastionpay_api/apigroup"
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
	fmt.Println(">api srv function jsonmessage")
	fmt.Println("		call api with json message")
	fmt.Println(">user srv function subuserkey jsonmessage")
	fmt.Println("		call user with json message")
	fmt.Println(">apitest srv function jsonmessage")
	fmt.Println("		call apitest with json message")
	fmt.Println(">apidoc [ver srv function]")
	fmt.Println("		list all apis path or by ver srv funcion")
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
			if len(argv) < 3 {
				fmt.Println("格式：api srv function message")
				continue
			}
			srv := argv[1]
			function := argv[2]
			message := ""
			if len(argv) > 3 {
				message = argv[3]
			}

			fmt.Println(message)

			var ack []byte
			err := gateway.RunApi("/api/v1/"+srv+"/"+function, []byte(message), &ack)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("err==", err)
			fmt.Println("ack==", string(ack))
		} else if argv[0] == "apitest"{
			if len(argv) < 3{
				fmt.Println("格式：apitest srv function message")
				continue
			}
			srv := argv[1]
			function := argv[2]
			message := ""
			if len(argv) > 3 {
				message = argv[3]
			}

			var ack []byte
			err := gateway.RunApiTest("/apitest/v1/"+srv+"/"+function, []byte(message), &ack)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("err==", err)
			fmt.Println("ack==", string(ack))
		} else if argv[0] == "user"{
			if len(argv) < 4 {
				fmt.Println("格式：user srv function subuserkey message")
				continue
			}
			srv := argv[1]
			function := argv[2]
			subUserKey := argv[3]
			message := ""
			if len(argv) > 4 {
				message = argv[4]
			}

			fmt.Println(message)

			var ack []byte
			err := gateway.RunUser("/user/v1/"+srv+"/"+function, subUserKey, []byte(message), &ack)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("err==", err)
			fmt.Println("ack==", string(ack))
		} else if argv[0] == "apidoc"{
			var ver string
			var srv string
			var function string
			if len(argv) > 1{
				ver = argv[1]
			}
			if len(argv) > 2{
				srv = argv[2]
			}
			if len(argv) > 3{
				function = argv[3]
			}

			if ver == "" && srv == ""{
				var index = 1
				apiAll := apigroup.ListApiGroup()
				for srv, apiGroup := range apiAll {
					fmt.Println("---------------------")
					fmt.Println("服务：")
					fmt.Println(srv)
					var subIndex = 1
					fmt.Println("---------------------")
					for _, apiProxy := range apiGroup{
						fmt.Println(index, ".", subIndex, ")", apiProxy.Help().Path(), " -- ", apiProxy.Help().Name)
						subIndex++
					}
					fmt.Println("---------------------")

					index++
				}

			} else if function == ""{
				var index = 1
				apiGroup, err := apigroup.ListApiGroupBySrv(ver, srv)
				if err != nil {
					fmt.Println("not find srv: ", err)
					continue
				}
				fmt.Println("---------------------")
				fmt.Println("服务：")
				fmt.Println(srv)
				for _, apiProxy := range apiGroup{
					fmt.Println("---------------------")
					fmt.Println(index, ")")
					fmt.Println("方法：")
					fmt.Println(apiProxy.Help().FuncName)
					fmt.Println("说明：")
					fmt.Println(apiProxy.Help().Name)
					fmt.Println("---------------------")
					index++
				}
			}else {
				apiProxy, err := apigroup.FindApiBySrvFunction(ver, srv, function)
				if err != nil {
					fmt.Println("not find function: ", err)
					continue
				}
				fmt.Println("---------------------")
				fmt.Println("版本：")
				fmt.Println(apiProxy.Help().VerName)
				fmt.Println("服务：")
				fmt.Println(apiProxy.Help().SrvName)
				fmt.Println("方法：")
				fmt.Println(apiProxy.Help().FuncName)
				fmt.Println("说明：")
				fmt.Println(apiProxy.Help().Name)
				fmt.Println("描述：")
				fmt.Println(apiProxy.Help().Description)
				fmt.Println("路径：")
				fmt.Println(apiProxy.Help().Path)

				fmt.Println("输入：")
				fmt.Println(apiProxy.Help().InputComment)

				fmt.Println("输出：")
				fmt.Println(apiProxy.Help().OutputComment)

				fmt.Println("示例：")
				fmt.Println(apiProxy.Help().Example)

				fmt.Println("---------------------")
			}
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