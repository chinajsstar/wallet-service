package main

import (
	"blockchain_server/service"
	"blockchain_server/types"
	"fmt"
	"blockchain_server/l4g"
	"time"
	"blockchain_server/conf"
	"blockchain_server/chains/btc"
	"context"
)

var (
	// RawTransaction hex
	//0200000001e1dab742abe7a575542010fea717b5aefeadfe1324bdbf39640450950675d52d0000000049483045022100e9a1e
	//5b6c5be14a33b01f9f4234531cde853a48dd8cd9a03e25be9d8e3553b8302206924c7390024d40df5088dad740b13e1731971
	//40122744ee781f02c4d60490e701feffffff0200ca9a3b000000001976a91401804fecf7a0980bded1f2cee7cac42e4306f5a
	//d88ac28196bee0000000017a914da785be00e58eefbf711e3286e4f06256dd7184587ca000000

	//seed_value	  :[3cfec70a88faa4c7a06aee3e869ff571c01673015610fb00830b16eff24eda3c],
	//extend_private:[tprv8ZgxMBicQKsPe9vm1tJZK4pKXBHEWMy6uQNCjGv8EFGH7CxHWLpcAkhN5dbsEo7B3KheFnrSmyjM5mzL8Bon985pkZdC6xWpbiEp3PFgbrX],
	//extend_public :[tpubD6NzVbkrYhZ4XcxYuXy9iUUS6CoAfhA1Uhxz1nxReX4fwhD48jeCMFKEFp2xPzmVLg3woBR1gnhHzbeL4JtvuVeByqvChjEfzC5dFYmGKpV],
	//account[0], crypt private key:7ddb84e9463b51a87ffdbc467c42fb4e01000000Qs1fCJ2olOq6GWrwz5jyR3C0, address:muSaiagrEsf41xkMT2jrVfS7z1x9umCdYb
	//account[1], crypt private key:1af172762746ae5722948e342a0abddb01000000z3vJeCL4HSCXltOFuRlh8Vy1, address:mmztibmBsXh3ezygzAMN52CtfCUwPWzXMy

	// virtual
	//-------≥≥≥≥≥≥≥≥≥ account[0], crypt private key:7ddb84e9463b51a87ffdbc467c42fb4e01000000Y9ZJED8gXxlonWxQ7SrsuDo0, address:muSaiagrEsf41xkMT2jrVfS7z1x9umCdYb
	//-------≥≥≥≥≥≥≥≥≥ account[1], crypt private key:1af172762746ae5722948e342a0abddb01000000LIuFf2sQeH65SD0ryO4mAxL1, address:mmztibmBsXh3ezygzAMN52CtfCUwPWzXMy
	//-------≥≥≥≥≥≥≥≥≥ account[2], crypt private key:52e5f1cb31fec078b4891c2c05a9661501000000O1qmt4M0TzMotp2AaKo3bH62, address:mfetgXxCwsh9U19v3Kv5j77FdgbchHjkzK
	//-------≥≥≥≥≥≥≥≥≥ account[3], crypt private key:8dcd7c7cbae3bb60d1d89f6d6aed5ff3010000008AICdUMlWb0tipCJnmKhwx33, address:mwBgJASmXyh3TF8PEkoTY8yePfi44P6aNA

	// real
	//--->Privatekey:[1/0], [key:cW8oWt451wQSC8MuQur7Pqmy71pPGXVwgX5CWirn4v24FxSNM6vc,address:muSaiagrEsf41xkMT2jrVfS7z1x9umCdYb]
	//--->Privatekey:[1/1], [key:cTwx1uw9eRXh13zpFt28Y6P2k1crKfBx31F6cg7Z4F2mGZWueN7J,address:mmztibmBsXh3ezygzAMN52CtfCUwPWzXMy]
	//--->Privatekey:[1/2], [key:cVpJKAxu8GhBLdyufUp2XcLJuxznm96qG2KA2REqBgKEyCARwCpX,address:mfetgXxCwsh9U19v3Kv5j77FdgbchHjkzK]
	//--->Privatekey:[1/3], [key:cPjkT4hJGRtj6MFuGKNg35BcMb9vvLg7t8Bo9NobZw3sncajy3KA,address:mwBgJASmXyh3TF8PEkoTY8yePfi44P6aNA]

	from_acc = &types.Account{
		PrivateKey: "7ddb84e9463b51a87ffdbc467c42fb4e01000000Qs1fCJ2olOq6GWrwz5jyR3C0",
		Address:	"muSaiagrEsf41xkMT2jrVfS7z1x9umCdYb" }
	to_acc = &types.Account{
		PrivateKey: "8dcd7c7cbae3bb60d1d89f6d6aed5ff3010000008AICdUMlWb0tipCJnmKhwx33",
		Address:	"mwBgJASmXyh3TF8PEkoTY8yePfi44P6aNA" }

	//watchonly = []types.Account{
	//	{"7ddb84e9463b51a87ffdbc467c42fb4e01000000wJNc61IJhWdoau5pCFHZSOU0","muSaiagrEsf41xkMT2jrVfS7z1x9umCdYb"},
	//	{"1af172762746ae5722948e342a0abddb01000000ES5fV9ShxQ71OfLKfhSafiR1","mmztibmBsXh3ezygzAMN52CtfCUwPWzXMy"},
	//	{"52e5f1cb31fec078b4891c2c05a96615010000008r2TlGCwDqd9eucreUVDMpm2","mfetgXxCwsh9U19v3Kv5j77FdgbchHjkzK"},
	//	{"8dcd7c7cbae3bb60d1d89f6d6aed5ff301000000uSNwOmVRsQRTgiK3ZdiHVQt3","mwBgJASmXyh3TF8PEkoTY8yePfi44P6aNA"},
	//}

	watchaddress = []string {
		"mmztibmBsXh3ezygzAMN52CtfCUwPWzXMy",
		"mfetgXxCwsh9U19v3Kv5j77FdgbchHjkzK",
        "mwBgJASmXyh3TF8PEkoTY8yePfi44P6aNA"}

	txid = "6f91ccf55f65806fbe161d64d3df5db8df9fbbc3108ac775345b560ff3834c5f"
	Coinname = types.Chain_bitcoin
	token string

	L4g = L4G.BuildL4g(types.Chain_bitcoin, "bitcoin")
)

func main() {
	L4g.Trace("-------------------bitcoin client sample start-------------------")
	clientManager := service.NewClientManager()
	client, err := btc.ClientInstance()

	if nil != err {
		fmt.Printf("create client:%s error:%s", Coinname, err.Error())
		return
	}

	// add client instance to manager
	clientManager.AddClient(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//*********批量创建账号示例*********/
	if false {
		accCmd := service.NewAccountCmd("message id", Coinname, 20)
		var accs []*types.Account
		accs, err = clientManager.NewAccounts(accCmd)
		for i, account := range accs {
			fmt.Printf("-------≥≥≥≥≥≥≥≥≥ account[%d], crypt private key:%s, address:%s\n",
				i, account.PrivateKey, account.Address)
		}
	}

	clientManager.Start()

	done_watchaddress := make(chan bool)
	done_sendTx := make(chan bool)

	i := 2
	if false {
		go testWatchAddress(ctx, clientManager, Coinname, nil, watchaddress, done_watchaddress)
	}else {i--}

	if true {
		go testSendTx(ctx, clientManager, from_acc.PrivateKey, to_acc.Address, Coinname,
			"", 0.1, done_sendTx)
	} else{i--}

	//testGetBalance(clientManager, from_acc.Address, token)

	for ; i>0; i-- {
		select {
		case <-done_sendTx: {
			L4g.Trace("SendTransaction done!")
		}
		case <-done_watchaddress: {
			L4g.Trace("Watch Address done!")
		}
		}
	}

	clientManager.Close()

	L4g.Trace("-------------------bitcoin client sample stop!-------------------")
	config.MainConfiger().Save()
	time.Sleep(time.Second)
}

func testWatchAddress(ctx context.Context, clientManager *service.ClientManager, coin string, token *string, addresslist []string, done chan bool) {
	defer L4g.Trace("exit watch address!!!")
	rcTxChannel := make(types.RechargeTxChannel)
	subscribe := clientManager.SubscribeTxRecharge(rcTxChannel)

	rcaCmd := service.NewRechargeAddressCmd("message id", Coinname, addresslist)
	clientManager.InsertRechargeAddress(rcaCmd)

	if false {
		addlist2 := []string{"0x8ce2af810e9f790e0a6d9f023ff3b7c35984aaad"}
		rcaCmd = service.NewRechargeAddressCmd("message id", Coinname, addlist2)
		clientManager.InsertRechargeAddress(rcaCmd)
	}

	watch_address_channel := make(chan bool)

	subCtx, cancel := context.WithCancel(ctx)
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		L4g.Trace("exit go routine of watch address!!")
		defer subscribe.Unsubscribe()
		defer close(channel)

		exit:
		for {
			select {
			case rct := <-channel:
				{
					if rct == nil {
						L4g.Trace("Watch Address channel is close!")
					} else {
						L4g.Trace("Recharge Transaction : cointype:%s, information:%s.", rct.Coin_name, rct.Tx.String())
						if rct.Tx.State == types.Tx_state_confirmed || rct.Tx.State == types.Tx_state_unconfirmed {
							watch_address_channel <- true
						}
					}
				}
			case <-ctx.Done():
				{
					fmt.Println("RechangeTx context done, because : ", ctx.Err())
					break exit
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

func testSendTx(ctx context.Context, clientManager *service.ClientManager,
	privatekey, to, coin string, token string, value float64, done chan bool) {

	txStateChannel := make(types.CmdTxChannel)
	subscribe := clientManager.SubscribeTxCmdState(txStateChannel)


	var (
		csend = 5
		cdone = 0
	)

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
						L4g.Trace("Transaction Command Channel is closed!")
						txok_channel <- false
					} else {
						L4g.Trace("Transaction state changed, transaction information:%s\n",
							cmdTx.Tx.String())
						if cmdTx.Error != nil {
							L4g.Trace("SendTx error: %s", cmdTx.Error.Message)
							cdone++
						} else {
							if cmdTx.Tx.State == types.Tx_state_confirmed ||
								cmdTx.Tx.State == types.Tx_state_unconfirmed {
									cdone++
							}
						}
						if cdone==csend {
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
	}(subCtx, txStateChannel)

	for i:=0; i<csend; i++ {
		txCmd, err := service.NewSendTxCmd(fmt.Sprintf("message id:%f", value),
			coin, privatekey, to, token, "", float64(i+1)/10)
		if err!=nil {
			L4g.Trace("err:%s", err.Error())
			return
		}
		clientManager.SendTx(txCmd)
	}


	select {
	case <-txok_channel:
		{
			cancel()
		}
	}
	done <- true
	return
}

func testGetBalance(manager *service.ClientManager, address string, tokenname string) {
	cmd_balance := service.NewQueryBalanceCmd("getbalance message id", Coinname, address, "")
	L4g.Trace("----------bitcoin get address balance---------")
	if balance, err := manager.GetBalance(context.TODO(), cmd_balance, nil); err == nil {
		L4g.Trace("address: %s balance: %d", address, balance)
	} else {
		L4g.Error("error : %s", err.Error())
	}
	L4g.Trace("----------bitcoin get address balance---------")
}
