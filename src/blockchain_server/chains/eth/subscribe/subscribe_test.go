package subscribe

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"testing"
	"blockchain_server/chains/eth"
	"time"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestSubscribe(t *testing.T) {
	//client, err := eth.NewHttpClient("http://127.0.0.1:8100")
	client, err := ethclient.Dial("ws://127.0.0.1:8500")
	if err != nil {
		fmt.Printf("connect to rpc server error:%v\n", err)
		return
	}
	fmt.Println("ethclient ok!!!")

	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second * 2)
	ch := make(chan bool)

	go func(ctx context.Context, ch chan bool){
		nid, err := client.NetworkID(ctx)
		if err!=nil {
			fmt.Println("Network id error:", err.Error())
			ch<- false
			return
		}
		fmt.Printf("ehereum client network id:%d\n", nid.Uint64())
		ch <- true
	}(ctx, ch)


	select {
	case <-ctx.Done():
		fmt.Println("ctx.done() with error:", ctx.Err())
	case isok := <-ch:
		if !isok {
			fmt.Println("cannot get net workid!!!! exit")
			return
		}
	}

	kstore := eth.DefualtKeyStore()
	if len(kstore.Accounts()) == 0 {
		fmt.Println("keystore account count is 0, exit")
		return
	} else {
		fmt.Printf("keystore has %d account\n", len(kstore.Accounts()))
		for i, ac := range kstore.Accounts() {
			fmt.Printf("account %d : 0x%x\n", i, ac.Address)
		}
	}

	fq := ethereum.FilterQuery{}

	fq.Addresses = make([]common.Address, 0)
	//fq.Addresses = append(fq.Addresses, kstore.Accounts()[0].Address)
	//fq.Addresses = append(fq.Addresses, kstore.Accounts()[1].Address)
	fq.FromBlock = nil
	fq.ToBlock   = nil

	ctx, cancelfunc = context.WithCancel(context.Background())
	defer cancelfunc()

	//logs_ch := make(chan types.Log, 50)


	if  false {
		ch_txHashString := make(chan string)
		subscription, err := client.SubscribePendingTransactions(ctx, ch_txHashString)
		if err!=nil {
			fmt.Printf("subscribefiler error:%s\n", err.Error())
			return
		}

		fmt.Printf("subscrpbe pending transactions ok, listenling....\n")
		tobreak := false
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("error happened:%v\n", ctx.Err())
				tobreak = true
			case tx_hashString := <-ch_txHashString:
				fmt.Printf("new pending transaction:%s\n", tx_hashString)
			}

			if tobreak {
				break
			}
		}
		subscription.Unsubscribe()
	}
	if  true {
		header_ch := make(chan *types.Header)
		subscription, err := client.SubscribeNewHead(ctx, true, header_ch)
		if err!=nil || nil==subscription {
			fmt.Printf("subscribefiler error:%s\n", err.Error())
			return
		}
		fmt.Printf("subcribe newblock ok, listenling...\n")
		tobreak := false
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("error happened:%v\n", ctx.Err())
				tobreak = true
			case header := <-header_ch:
				fmt.Printf("new block header : %s\n", header.String())
			}

			if tobreak {
				break
			}
		}
	}



}
