package service

import (
	"testing"
	"blockchain_server/types"
	"blockchain_server/chains/eth"
	"fmt"
	"context"
	l4g "github.com/alecthomas/log4go"
	"time"
)

func TestNetCmdSvr(t *testing.T) {
	rctChannel := make(types.RechargeTxChannel)
	clientManager := &ClientManager{}
	client, err := eth.NewClient()

	if nil!=err {
		fmt.Printf("create client:%s error:%s", types.Chain_eth, err.Error() )
		return
	}

	// add client instance to manager
	clientManager.AddClient(client)


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/*********批量创建账号示例*********/
	accCmd := types.NewAccountCmd("message id", types.Chain_eth, 10)
	var accs []*types.Account
	accs, err = clientManager.NewAccounts(accCmd)
	for i, account := range accs {
		fmt.Printf("account[%d], crypt private key:%s, address:%s\n",
			i, account.PrivateKey, account.Address)
	}

	/*********添加监控地址示例*********/
	addresses := []string{accs[0].Address, accs[1].Address}
	rcaCmd := types.NewRechargeAddressCmd("message id", types.Chain_eth, addresses)
	clientManager.InsertRechargeAddress(rcaCmd)


	/*********监控提币交易的channel*********/
	txStateChannel := make(types.CmdTxChannel)

	// 创建并发送Transaction, 订阅只需要调用一次, 所有的Send的交易都会通过这个订阅channel传回来
	clientManager.SubscribeTxCmdState(txStateChannel)

	ctx2, _ := context.WithCancel(ctx)
	txok_channel := make(chan bool)
	go func(ctx2 context.Context, txstateChannel types.CmdTxChannel) {
		close := false
		for !close {
			select {
			case cmdTx := <-txStateChannel:{
				fmt.Printf("Transaction state changed, transaction information:%s\n",
					cmdTx.Tx.String())

				if cmdTx.Tx.State == types.Tx_state_confirmed {
					l4g.Trace("Transaction is confirmed! success!!!")
					txok_channel <- true
				}

				if cmdTx.Tx.State == types.Tx_state_unconfirmed {
					l4g.Trace("Transaction is unconfirmed! failed!!!!")
					txok_channel <- false
				}

			}
			case <-ctx.Done():{
				close = true
			}
		}
		}
	}(ctx2, txStateChannel)


	/*********执行提币命令*********/
	txCmd := types.NewTxCmd("message id", types.Chain_eth, accs[0].PrivateKey, accs[1].Address, 10000)
	clientManager.SendTx(txCmd)


	/*********创建监控充币地址channael*********/
	watch_address_channel := make(chan bool)
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		exit := false
		for !exit {
			select {
			case rct := <-channel:{
				fmt.Printf("Recharge Transaction : cointype:%s, information:%s.", rct.Coin_name, rct.Tx.String())
				watch_address_channel <- true
			}
			case <-ctx.Done():{
				fmt.Println("RechangeTx context done, because : ", ctx.Err())
				exit = true
			}
			}
		}
	}(ctx, rctChannel)

	/*********开启服务!!!!!*********/
	rcTxChannel := make(types.RechargeTxChannel)
	clientManager.SubscribeRechargeTx(rcTxChannel)
	clientManager.Start()


	okcount := 0
	select {
	case <-watch_address_channel :{
		okcount++
		l4g.Trace("watching address gorouine already exited!")
		if okcount==2 {
			ctx.Done()
		}
	}
	case   <-txok_channel :{
		okcount++
		l4g.Trace("transaction gorouine already exited!")
		if okcount==2 {
			ctx.Done()
		}
	}
	}

	clientManager.Close()

	l4g.Trace("exit main!")

	time.Sleep(1 * time.Second)
}
