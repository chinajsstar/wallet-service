package main

import (
	"context"
	"time"
	"fmt"
	"blockchain_server/chains/btc"
	"strings"
	"api_router/base/utils"
	"github.com/btcsuite/btcutil"
	"strconv"
)

func main()  {
	cc, err := btc.ClientInstance("localhost:18444", "henly", "henly123456", ":8076")
	if err != nil {
		fmt.Println("#ClientInstance failed:", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	cc.Start(ctx)

	time.Sleep(time.Second*1)
	for ; ;  {
		fmt.Println("Input 'quit' to quit...")
		var input string
		input = utils.ScanLine()

		argv := strings.Split(input, " ")

		if argv[0] == "quit" {
			cancel()
			break;
		}else if argv[0] == "newaddr" {
			if len(argv) < 2 {
				fmt.Println("err:", "2 argv")
				continue
			}
			addr, err := cc.GetNewAddress(argv[1])
			if err != nil {
				fmt.Println("err:", err)
				continue
			}

			fmt.Println(addr.String())
			fmt.Println(addr.EncodeAddress())
			fmt.Println(addr.ScriptAddress())
		}else if argv[0] == "newaddr2" {
			addr, err := cc.NewAccount()
			if err != nil {
				fmt.Println("err:", err)
				continue
			}

			fmt.Println("add:", addr)
		}else if argv[0] == "sendtx" {
			if len(argv) < 3 {
				fmt.Println("err:", "3 argv")
				continue
			}
			dd, _ := strconv.ParseFloat(argv[2], 10)
			amount, _ := btcutil.NewAmount(float64(dd))

			addr, _ := btcutil.DecodeAddress(argv[1], nil)
			ret, err := cc.SendToAddress(addr, amount)
			if err != nil {
				fmt.Println("sendtoaddress failed: ", err)
			}

			fmt.Println("ret:", ret)
		}
	}

	fmt.Println("stop to quit...")
	cc.Stop()

	fmt.Println("wait to quit...")
	cc.WaitForShutdown()
}
