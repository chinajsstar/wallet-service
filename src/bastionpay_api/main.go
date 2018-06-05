package main

import (
	"fmt"
	"strings"
	"bastionpay_api/utils"
	"bastionpay_api/gateway"
	"bastionpay_api/apigroup"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"bastionpay_api/api"
)

func userUsage()  {
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
	fmt.Println(">user subuserkey srv function jsonmessage")
	fmt.Println("		call user with json message")
	fmt.Println(">admin subuserkey srv function jsonmessage")
	fmt.Println("		call admin with json message")
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

	userUsage()

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
			userStartHttpServer(port)

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
				fmt.Println("格式：user subuserkey srv function message")
				continue
			}
			subUserKey := argv[1]
			srv := argv[2]
			function := argv[3]

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
		} else if argv[0] == "admin"{
			if len(argv) < 4 {
				fmt.Println("格式：admin subuserkey srv function message")
				continue
			}
			subUserKey := argv[1]
			srv := argv[2]
			function := argv[3]

			message := ""
			if len(argv) > 4 {
				message = argv[4]
			}

			fmt.Println(message)

			var ack []byte
			err := gateway.RunAdmin("/admin/v1/"+srv+"/"+function, subUserKey, []byte(message), &ack)
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
				fmt.Println(apiProxy.Help().Path())

				fmt.Println("输入：")
				fmt.Println(utils.FormatComment(apiProxy.Help().Input))

				fmt.Println("输出：")
				fmt.Println(utils.FormatComment(apiProxy.Help().Output))

				fmt.Println("示例：")
				fmt.Println(utils.FormatSample(apiProxy.Help().Input))

				fmt.Println("---------------------")
			}
		} else if argv[0] == "decode"{
			if len(argv) < 3 {
				fmt.Println("decode message signature")
				//continue
			}
			ud := api.UserData{}
			ud.Message = `Erf4ULyC05ayM0jLrncDxzDDqmCz1bDKzUerFerci0rNRf7CkWQb1ZRijEE5LZIrvIiYzJe3oEQXL/OwcAA0Mp/gWlpzF7hwWaLwGYWGWt1bdW/yEd3HbHlUUoS0rxi17fJ2EkvBsGkSaHVLm3YGss9TF3udrFQJDqh0M6yKAYeMBr8sARQc4GIosVMyctBTdOwI/GYzisvKrc4N+jE5YtORP9zkXybDvRluyAW2HDKZ9FYNJFxFXu/OWaTCJbI5Z8oati/DhM/v749kxXapIIoPBuKWIAtl9IbfdTyZdXoMcBzVF4r5eS0MU/Xs4KIK9gKdOBY2wmyWuWpcUjk+57ERfxDoFP5DHc8DGVfo+4gFCsmYxp+qTDSJW2M9xmH75QofpyHgLeUcHkKipX4SMVSk/FYhaVKBJ+FP+BXpr7Xl/n9aZkzTK8/qpgF4sqnlGxpNE6jl/nBgVA1GLqTaG+zkNdMWY7bpSGR5nYc1Kb6MUnvLgBwRTiwspWA57E16gUvx9SJLlgkLbHUKZtuArpOovaJWA+TR3IzC9QhiloCWuyi03XLJVDvXrPqm/6APzL+4XwTEE98NJR9SgQzBHTdcqgwiavS/eq5MtPQFt9g388j80mc/vsvP57Odqg7BbE1DKZyw17RdfkyPwYfyJZQTEymUgaH0uzq7wp46KpE=`
			ud.Signature = `h/SYdhV8srfchOHVN44Vo/v2V5nqmG26YaLz76b72P93VSzz66vwuNObANQpHlXjRMDW49IokkBIiqB06/GUEidVNrg3A+4tQILSWSZlOkKFHAxC7VctIAlllkxqYtIVyILPB5e9LEHXDVUt5coLoVpzbkDg7uihoBMNiAtQxn89H7AduRVrEtTCnqhmbbsaS3yMDqsx/ArODXYUWU8nxCglYQIVI4w7FBu4S4Wl1pe353FzvNE07tG5/x8htt4sRXf5btYj3muvKq1Ch6PuJgIV4hSc3dgKzMA3jOG4PJhZWyFQCT0NkjOHoBO1RtvuT4wbi1IUNRtEVvCOySxxzQ==`
			a, err := gateway.Decryption(&ud)
			fmt.Println(string(a))
			fmt.Println(err)
		} else {
			userUsage()
		}
	}
}

// 处理推送
func userHandlePush(w http.ResponseWriter, req *http.Request) {
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
func userStartHttpServer(port string) error {
	fmt.Println("Start http server on: ", port)

	http.Handle("/walletcb", http.HandlerFunc(userHandlePush))
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