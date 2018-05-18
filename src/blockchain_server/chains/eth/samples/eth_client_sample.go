package main

import (
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	"context"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"time"
)

var (
	// 0x04c8972b7274f98afa6bc529ac0ecb0c4669b09d253bb856d6990dae5ef4c11921b6fc0b31dc4f94c9b38c6311a390f8fb61e0b8b4a6aee5cda00bce50181f3256bc03a6d355d2ed70e0fb09485cb6fff01e7549b138f88fa9729f676fef5ccd75177f89052b0aa3d6b224b79af3bf194c582826e01ebb5a0e987b6b8d8babcea0a8c461d554b5999bf70ad644918f7129
	// 0x0B9CB828028306470994a09b584bC7BAd196131a
	bank_account = &types.Account{
		"0x04bac76a1adee95417514a449efe0f1e6ccfb856b1f077a2b61dd0bdf8fd71defb4b9761d27bd0bf3f3f80094844edb19b55c54d159254d4b4230a88a2f19202093df5687bc03eff86875baf1fa45710073b3496fb8a4edcee9f54af9213ec41b5d0380454435ec4361892b1ae8dca64f9343b2fd46a492b2bd9623b1f23757b5398de5b33606835e68824c9bd0d8f516c",
		"0x5c0ebCBAA341147c1B98D097BeD99356f8B8340F"}
	//to_account = &types.Account{
	//	"0x046c3de0582b9eaaa0ad9f04d26227ec36dc3bcb128f3e2739e3cc4c5fd92d1e177731e83c19366d9332fe51aea0105becd96b6100122e38fcc19b8edabffe4ecafa980aec412a949189a49c63311a4d376cd5d9db98a2e5a7ce6ab2db7ba13df91d5a9229bff158d564dacfa65ffa3f74580cdfd060d7b94d265330d3edeea3172a59b00d2a4ed77bd95ab7e2bf704ce9",
	//	"0xD35F8FF353d3b1B08305209f5CEb53333134D381"}

	accounts = []*types.Account{
		&types.Account{"0x040f13f1d287e31a7f5703cf14a94adcfc163f8116569e364c039848a13e0b8581dd9b4e8f7952031ad634262bec3e289b6aa30125984a19ea5d48015e1221a67f8f990e5ed8a06dcdc14240952ae510daeb4626f57a740546be034f01a7163e93e76712552e1a0f23a221e04b85836befd646a268331e0907654464ae21c2c01bdf9d38e98a1618df15e4c0b75e95be5f",
			"0xc23f8B98Ba550Bc1f1DBF58B8D93a444020dB3cc"},
		&types.Account{"0x04b6964cf44a4c2a6c5b84df3738fb68bec3f34d603d6699190c7407554924eea5a6ca83bd4b4361b54940b96806b1492e3fca9d1b597f5329bd40d1304988fcb5d7cc6809c27a7041c82172e7bed0a3b90678d1dc7625328a4dbceb9376fb50d599505b4bbb2dcfeba57059f9d9e6fcc070af90e9c56ca200ecff8e8d972346a1a868a9c91d7f148e5be8e0eba99d1c90",
			"0x969eBf677d1E9f31fd3a2Aee92b6f4E85De70545"},
			&types.Account{"0x04ba927621c916e0c5e8b619cf6da9005795fc70378042b4494e7fd1201389fb1e19ee7b68615daf590e614112da4e1b028bc40fb48bd6d6b26a5c0efb1a06f096d5e77c000a3db17327c150b44ddd7fc8112b0763a8059c8f5c2742b14d9be91977883390e7a9a2865ba647420c3bb4d60c0c3355aa88c6803e83db62d1325f22396b6c2177cfe549d3c908cc32ccfe84",
			"0x0B9CB828028306470994a09b584bC7BAd196131a"},
	}

	callerAcc      = bank_account
	tokenReciptAcc = accounts[0]
	tokenOwnerAcc  = accounts[1]

	to_account = accounts[0]
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
	if false {
		accCmd := service.NewAccountCmd("message id", types.Chain_eth, 2)
		var accs []*types.Account
		accs, err = clientManager.NewAccounts(accCmd)
		for i, account := range accs {
			fmt.Printf("account[%d], Account{\"%s\",\n\"%s\"}\n",
				i, account.PrivateKey, account.Address)
		}
	}
	clientManager.Start()

	done_watchaddress := make(chan bool)
	done_sendTx := make(chan bool)

	token := "ZToken"
	i := 0
	if false {
		go testWatchAddress(ctx, clientManager, types.Chain_eth, nil, []string{to_account.Address, bank_account.Address}, done_watchaddress)
	} else {i++}

	if true {
		go testSendTokenTx(
			ctx,
			clientManager,
			callerAcc.PrivateKey,
			tokenOwnerAcc.PrivateKey,
			tokenReciptAcc.Address,
			types.Chain_eth,
			token, 1.2,
			done_sendTx)
	} else {i++}

	testGetBalance(clientManager, bank_account.Address, token)

	for ; i < 2; i++ {
		select {
		case <-done_sendTx:
			{
				l4g.Trace("SendTransaction done!")
			}
		case <-done_watchaddress:
			{
				l4g.Trace("Watch AddressOfContract done!")
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
						l4g.Trace("Watch AddressOfContract channel is close!")
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

func testSendTokenTx(ctx context.Context, clientManager *service.ClientManager,
	callFromKey, tokenOwnerKey, tokenTo string,
	coin string, token string, value float64, done chan bool) {

	txCmd, err := service.NewSendTxCmd("MessageID:0000000011",
		coin, callFromKey, tokenTo, token,
		tokenOwnerKey, value)

	if err != nil {
		l4g.Trace("CreateSendTxCmd faild, message:%s", err.Error())
		done <- false
		return
	}
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

func testGetBalance(manager *service.ClientManager, address string, tokenSymbol string) {
	cmd_balance := service.NewQueryBalanceCmd("getbalance message id", types.Chain_eth, address, "")
	cmd_balance_token := service.NewQueryBalanceCmd("getbalance message id", types.Chain_eth, address, tokenSymbol)

	l4g.Trace("------------------")
	if balance, err := manager.GetBalance(context.TODO(), cmd_balance, nil); err == nil {
		l4g.Trace("balance(%s) = %d", address, balance)
	} else {
		l4g.Error("error : %s", err.Error())
	}

	l4g.Trace("------------------")
	if balance, err := manager.GetBalance(context.TODO(), cmd_balance_token, nil); err == nil {
		l4g.Trace("%s balance(%s) = %d", tokenSymbol, address, balance)
	} else {
		l4g.Error("error : %s", err.Error())
	}

}
