package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"math"
	"math/big"
)

type Zl_Client struct {
	eth_client *ethclient.Client
	address    string
}

func (zl_client *Zl_Client) getBalance() (float64, error) {
	return 0, nil
}

func connectToRpc() (*ethclient.Client, error) {
	client, err := rpc.Dial("http://127.0.0.1:8100")
	if err != nil {
		return nil, err
	}
	conn := ethclient.NewClient(client)
	return conn, nil
}

func GetBalance(address string) float64 {
	client, err := connectToRpc()
	if err != nil {
		panic(err.Error())
	}

	balance, err := client.BalanceAt(context.TODO(), common.HexToAddress(address), nil)

	var balanceVal float64
	if err != nil {
		panic(err)
		fmt.Println("error subscribe:", err.Error())
		balanceVal = 0.0
	} else {
		fmt.Printf("Big Int: %v\n", balance)
		fbalance := new(big.Float).SetInt(balance)
		fbalance.Mul(fbalance, big.NewFloat(math.Pow(10, -18)))
		balanceVal, _ = fbalance.Float64()
	}
	return balanceVal
}

//func getBalance() {
//	client, err := ethclient.Dial("http://127.0.0.1:8100")
//
//	if err != nil {
//		fmt.Printf("%v", err)
//		os.Exit(1)
//
//		client.BalanceAt
//	}
//}

func main() {
	log.Fatal("error!!!!!!")
	fmt.Println("did not exit!!!!!")

	return

	var balance float64
	var account_address = "0xa40f6bf261914447987959ce26880d22eddf7dc6"
	balance = GetBalance(account_address)

	fmt.Println("address:", account_address, " balance:", balance)
}
