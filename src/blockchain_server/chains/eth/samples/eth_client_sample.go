package main

import (
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	"context"
	"fmt"
	"blockchain_server/l4g"
	"time"
	"blockchain_server/conf"
)

var (
	// 0x04c8972b7274f98afa6bc529ac0ecb0c4669b09d253bb856d6990dae5ef4c11921b6fc0b31dc4f94c9b38c6311a390f8fb61e0b8b4a6aee5cda00bce50181f3256bc03a6d355d2ed70e0fb09485cb6fff01e7549b138f88fa9729f676fef5ccd75177f89052b0aa3d6b224b79af3bf194c582826e01ebb5a0e987b6b8d8babcea0a8c461d554b5999bf70ad644918f7129
	// 0x0B9CB828028306470994a09b584bC7BAd196131a

	// 0x046c3de0582b9eaaa0ad9f04d26227ec36dc3bcb128f3e2739e3cc4c5fd92d1e177731e83c19366d9332fe51aea0105becd96b6100122e38fcc19b8edabffe4ecafa980aec412a949189a49c63311a4d376cd5d9db98a2e5a7ce6ab2db7ba13df91d5a9229bff158d564dacfa65ffa3f74580cdfd060d7b94d265330d3edeea3172a59b00d2a4ed77bd95ab7e2bf704ce9",
	// 0xD35F8FF353d3b1B08305209f5CEb53333134D381"}

	//bank_account = &types.Account{
	//	"0x04bac76a1adee95417514a449efe0f1e6ccfb856b1f077a2b61dd0bdf8fd71defb4b9761d27bd0bf3f3f80094844edb19b55c54d159254d4b4230a88a2f19202093df5687bc03eff86875baf1fa45710073b3496fb8a4edcee9f54af9213ec41b5d0380454435ec4361892b1ae8dca64f9343b2fd46a492b2bd9623b1f23757b5398de5b33606835e68824c9bd0d8f516c", "0x5c0ebCBAA341147c1B98D097BeD99356f8B8340F"}

	default_accs = []*types.Account{
		&types.Account{"0x04c7a44b0ddfe6026d632e704a7ca10551a9ba659ed1c9e2593162a385b11b9a0a69b2f7050a53bd4e1665e4a07a018e91e8ec4d17edc002db63c2df4ab31877e3cca2f6b04e0923939fd62df40468191460e20114fedf8b783b23e31164d5e75475008bebbf01c3b04d5e62739b733c10895ab5b21bdd997d0c13ce420c9b4b01969633e513d57a500146ba8fa843474f", "0x69F7337302Aec7F6ae7915db3f31da865214e771"},
		&types.Account{"0x04862a4bc9eb67099781d005a3e39c7e1f7d1eaf8e9cbcc239b4191203f4380bff4745062810a35ef34c5a7eb07b163dd145347487b524f461478793f658ce2b0ad560d3ebc9d379edf8fec31dd3b61af81674cdf7867778a2d705de04fa8d31eddedcf182678c1ece565f5e4adf455a02fb24e0a1d9b15353e5bc7f1f13512e414960db1c4ca87a526a798faa0802ec65", "0x969eBf677d1E9f31fd3a2Aee92b6f4E85De70545"},
		&types.Account{"0x0478303ffb7fc6841ba5fac2e54fc220a4a0dec0e7ab0b44d1a4171cbdd34e111df7a7e94baf1ed6412083c524afb06f87e5f596b0090e5a242279cc16faa6bd2956862c1b46c84c26cc05f45d47abe19259fe7d0e48253d34df8fa598976e4e8e5ca6fcfec0ac4064a105dccc950010ca0aaa8a9b44c559e87c819d41d06060db07cdef25ade94c3eb9791dbe126ae394", "0x43a957847D88c019C621847B1bD1B917741DbE3a"}}

	bank_account   = default_accs[0]
	token_owner    = default_accs[0]
	caller         = default_accs[0]

	to_account     = default_accs[1]
	token_receiver = default_accs[2]
)

func main() {
	clientManager := service.NewClientManager()

	client, err := eth.ClientInstance()

	if nil != err {
		L4g.Error("create client:%s error:%s", types.Chain_eth, err.Error())
		L4g.Close()
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

	token := "ZTK"
	i := 0
	if false {
		go watchRechargeTxByAddress(ctx, clientManager, types.Chain_eth,
			[]string{default_accs[1].Address, default_accs[2].Address},
			done_watchaddress)
	} else {
		i++
	}

	if true {
		go send_tokenTx(
			ctx,
			clientManager,
			caller.PrivateKey,
			token_owner.PrivateKey,
			token_receiver.Address,
			types.Chain_eth,
			token, float64(i+1),
			done_sendTx)
	} else {
		i++
	}

	get_balance(clientManager, bank_account.Address, token)

	for ; i < 2; i++ {
		select {
		case <-done_sendTx:
			{
				L4g.Trace("SendTransaction done!")
			}
		case <-done_watchaddress:
			{
				L4g.Trace("Watch AddressOfContract done!")
			}
		}
	}

	clientManager.Close()

	L4g.Trace("exit main!")

	time.Sleep(1 * time.Second)
}

func watchRechargeTxByAddress(ctx context.Context, clientManager *service.ClientManager, coin string, addresslist []string, done chan bool) {
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
		defer close(channel)
		defer subscribe.Unsubscribe()

		exit := false
		for !exit {
			select {
			case rct := <-channel:
				{
					if rct == nil {
						L4g.Trace("Watch AddressOfContract channel is close!")
					} else {
						L4g.Trace("Recharge Transaction : cointype:%s, information:%s.",
							rct.Coin_name, rct.Tx.String())

						config.MainConfiger().Save()
						if (rct.Tx.State == types.Tx_state_confirmed || rct.Tx.State == types.Tx_state_unconfirmed) && false {
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

func send_tokenTx(ctx context.Context, clientManager *service.ClientManager,
	callFromKey, tokenOwnerKey, tokenTo string,
	coin string, token string, value float64, done chan bool) {

	txStateChannel := make(types.CmdTxChannel)
	subscribe := clientManager.SubscribeTxCmdState(txStateChannel)
	txok_channel := make(chan bool)
	subCtx, cancel := context.WithCancel(ctx)

	csend := 0
	cdone := 0

	var makeAndSend = func(coin, callfromkey, tokenTo, token, tokenOwnerkey string, value float64) {
		txCmd, err := service.NewSendTxCmd(fmt.Sprintf("MsgId:%f", value),
			coin, callFromKey, tokenTo, token,
			tokenOwnerKey, value)
		if err != nil {
			L4g.Trace("CreateSendTxCmd faild, message:%s", err.Error())
			L4G.Close("all")
			return
		}
		clientManager.SendTx(txCmd)
	}

	var waiting = func(ctx context.Context, txstateChannel types.CmdTxChannel) {
		defer subscribe.Unsubscribe()
		defer close(txstateChannel)
		close := false
		for !close {
			select {
			case cmdTx := <-txStateChannel:
				{
					if cmdTx == nil {
						L4g.Trace("Transaction Command Channel is closed!")
						cdone++
					} else {
						L4g.Trace("Transaction state changed, hash=%s", cmdTx.Tx.Tx_hash)
						if cmdTx.Tx.State == types.Tx_state_confirmed ||
							cmdTx.Tx.State == types.Tx_state_unconfirmed {
							cdone++
						}

						if cdone == csend {
							txok_channel <- true
						}
					}
				}
			case <-ctx.Done():
				{
					close = true
				}
			}
		}
	}

	for i := 0; i < 1; i++ {
		go makeAndSend(coin, callFromKey, tokenTo, token, tokenOwnerKey, float64(i+1))
		csend++
	}

	go waiting(subCtx, txStateChannel)

	select {
	case <-txok_channel:
		{
			cancel()
		}
	}

	done <- true
}

func get_balance(manager *service.ClientManager, address string, tokenSymbol string) {
	cmd_balance := service.NewQueryBalanceCmd("getbalance message id", types.Chain_eth, address, "")
	cmd_balance_token := service.NewQueryBalanceCmd("getbalance message id", types.Chain_eth, address, tokenSymbol)

	L4g.Trace("------------------")
	if balance, err := manager.GetBalance(context.TODO(), cmd_balance, nil); err == nil {
		L4g.Trace("balance(%s) = %f", address, balance)
	} else {
		L4g.Error("error : %s", err.Error())
	}

	L4g.Trace("------------------")
	if balance, err := manager.GetBalance(context.TODO(), cmd_balance_token, nil); err == nil {
		L4g.Trace("%s balance(%s) = %f", tokenSymbol, address, balance)
	} else {
		L4g.Error("error : %s", err.Error())
	}
	L4g.Trace("------------------")
}
