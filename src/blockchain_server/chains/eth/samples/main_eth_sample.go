package main

import (
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	"context"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"time"
	"math"
)

var (
	bank_account = &types.Account{
		"0x04a7983149693e7d26571801a9cdc8c6cb1e76a7666f65a8cd8c7af343abd46440469e3dc00faabc9db96764e6063002452f36227b8f6fe262c2c2387116d20b2cb63e0987cd14e8b4156c8ee0149d2137086b23ef1bf691555990d0bf036bd7ba5617ad2179a296e23e9749461d3bb6d97fd645b52fdaa83245480f0a8c77b4650f47fb7de97fee43d3db544ec8e55120",
		"0xadFF7cAE7b43B9990789dFD41791218dd75307Fe"}

	to_account = &types.Account{
		"0x046c3de0582b9eaaa0ad9f04d26227ec36dc3bcb128f3e2739e3cc4c5fd92d1e177731e83c19366d9332fe51aea0105becd96b6100122e38fcc19b8edabffe4ecafa980aec412a949189a49c63311a4d376cd5d9db98a2e5a7ce6ab2db7ba13df91d5a9229bff158d564dacfa65ffa3f74580cdfd060d7b94d265330d3edeea3172a59b00d2a4ed77bd95ab7e2bf704ce9",
		"0xD35F8FF353d3b1B08305209f5CEb53333134D381"}
)
func main() {
	clientManager := service.NewClientManager()

	client, err := eth.ClientInstance()

	if nil != err {
		fmt.Printf("create client:%s error:%s", types.Chain_eth, err.Error())
		return
	}

	// add client instance to manager
	clientManager.AddClient(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/*********批量创建账号示例*********/
	if true {
		accCmd := service.NewAccountCmd("message id", types.Chain_eth, 2)
		var accs []*types.Account
		accs, err = clientManager.NewAccounts(accCmd)
		for i, account := range accs {
			fmt.Printf("account[%d], crypt private key:%s, address:%s\n",
				i, account.PrivateKey, account.Address)
		}
	}
	clientManager.Start()

	done_watchaddress := make(chan bool)
	done_sendTx := make(chan bool)

	token := "ZToken"
	if true {
		go testWatchAddress(ctx, clientManager, types.Chain_eth, nil, []string{to_account.Address, bank_account.Address}, done_watchaddress)
	}

	if false {
		go testSendTokenTx(ctx, clientManager, bank_account.PrivateKey,to_account.Address, types.Chain_eth,
			&token, 100 * uint64(math.Pow10(8)), done_sendTx)
	}

	testGetBalance(clientManager, bank_account.Address, token)

	for i:=0; i<2; i++ {
		select {
		case <-done_sendTx: {
			l4g.Trace("SendTransaction done!")
		}
		case <-done_watchaddress: {
			l4g.Trace("Watch Address done!")
		}
		}
	}

	clientManager.Close()

	l4g.Trace("exit main!")

	time.Sleep(1 * time.Second)
}

func testWatchAddress(ctx context.Context, clientManager *service.ClientManager, coin string, token *string, addresslist []string, done chan bool) {
	rcTxChannel := make(types.RechargeTxChannel)
	subscribe := clientManager.SubscribeTxRecharge(rcTxChannel)

	rcaCmd := service.NewRechargeAddressCmd("message id", types.Chain_eth, addresslist)
	clientManager.InsertRechargeAddress(rcaCmd)

	if true {
		addlist2 := []string{"0x8ce2af810e9f790e0a6d9f023ff3b7c35984aaad"}
		rcaCmd = service.NewRechargeAddressCmd("message id", types.Chain_eth, addlist2)
		clientManager.InsertRechargeAddress(rcaCmd)
	}

	watch_address_channel := make(chan bool)

	subCtx, cancel := context.WithCancel(ctx)
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		defer subscribe.Unsubscribe()
		defer close(channel)

		exit := false
		for !exit {
			select {
			case rct := <-channel:
				{
					if rct == nil {
						l4g.Trace("Watch Address channel is close!")
					} else {
						l4g.Trace("Recharge Transaction : cointype:%s, information:%s.", rct.Coin_name, rct.Tx.String())
						if rct.Tx.State == types.Tx_state_confirmed || rct.Tx.State == types.Tx_state_unconfirmed {
							watch_address_channel <- true
						}
					}
				}
			case <-ctx.Done():
				{
					fmt.Println("RechangeTx context done, because : ", ctx.Err())
					exit = true
				}
			}
		}
	}(subCtx, rcTxChannel)

	select {
	case <-watch_address_channel:
		cancel()
	}

	done <- true
}

func testSendTokenTx(ctx context.Context, clientManager *service.ClientManager, privatekey, to, coin string,
	token *string, value uint64, done chan bool) {
	txCmd := service.NewSendTxCmd("message id", coin, privatekey, to, token, value)
	clientManager.SendTx(txCmd)

	/*********监控提币交易的channel*********/
	txStateChannel := make(types.CmdTxChannel)

	// 创建并发送Transaction, 订阅只需要调用一次, 所有的Send的交易都会通过这个订阅channel传回来
	subscribe := clientManager.SubscribeTxCmdState(txStateChannel)

	txok_channel := make(chan bool)

	subCtx, cancel := context.WithCancel(ctx)

	go func(ctx context.Context, txstateChannel types.CmdTxChannel) {
		defer subscribe.Unsubscribe()
		defer close(txstateChannel)
		close := false
		for !close {
			select {
			case cmdTx := <-txStateChannel:
				{
					if cmdTx == nil {
						l4g.Trace("Transaction Command Channel is closed!")
						txok_channel <- false
					} else {
						l4g.Trace("Transaction state changed, transaction information:%s\n",
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
				}
			case <-ctx.Done():
				{
					close = true
				}
			}
		}
	}(subCtx, txStateChannel)

	select {
	case <-txok_channel:
		{
			cancel()
		}
	}
	done <- true
}

func testGetBalance(manager *service.ClientManager, address string, tokenname string) {
	cmd_balance := service.NewQueryBalanceCmd("getbalance message id", types.Chain_eth, address, nil)
	cmd_balance_token := service.NewQueryBalanceCmd("getbalance message id", types.Chain_eth, address, &tokenname)

	l4g.Trace("------------------")
	if balance, err := manager.GetBalance(context.TODO(), cmd_balance, nil); err == nil {
		l4g.Trace("balance(%s) = %d", address, balance)
	} else {
		l4g.Error("error : %s", err.Error())
	}

	l4g.Trace("------------------")
	if balance, err := manager.GetBalance(context.TODO(), cmd_balance_token, nil); err == nil {
		l4g.Trace("%s balance(%s) = %d", tokenname, address, balance)
	} else {
		l4g.Error("error : %s", err.Error())
	}

}
