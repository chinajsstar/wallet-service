package service

import (
	"testing"
	"blockchain_server/types"
	"blockchain_server/chains/eth"
	"fmt"
	"context"
	L4g "blockchain_server/l4g"
	"time"
)

var (
	tmp_account = &types.Account{
		"0x04e2b6c9bfeacd4880d99790a03a3db4ad8d87c82bb7d72711b277a9a03e49743077f3ae6d0d40e6bc04eceba67c2b3ec670b22b30d57f9d6c42779a05fba097536c412af73be02d1642aecea9fa7082db301e41d1c3c2686a6a21ca431e7e8605f761d8e12d61ca77605b31d707abc3f17bc4a28f4939f352f283a48ed77fc274b039590cc2c43ef739bd3ea13e491316",
		"0x54b2e44d40d3df64e38487dd4e145b3e6ae25927"}
	tmp_toaddress = "0x0c14120e179f7dc6571467448fb3a7f7b14f889d"
)

func TestNetCmdSvr(t *testing.T) {
	clientManager := NewClientManager()

	client, err := eth.ClientInstance()

	if nil!=err {
		fmt.Printf("create client:%s error:%s", types.Chain_eth, err.Error() )
		return
	}

	// add client instance to manager
	clientManager.AddClient(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/*********批量创建账号示例*********/
	//accCmd := NewAccountCmd("message id", types.Chain_eth, 10)
	//var accs []*types.Account
	//accs, err = clientManager.NewAccounts(accCmd)
	//for i, account := range accs {
	//	fmt.Printf("account[%d], crypt private key:%s, address:%s\n",
	//		i, account.PrivateKey, account.Address)
	//}
	clientManager.Start()

	done_watchaddress := make(chan bool)
	done_sendTx := make(chan bool)

	//go testWatchAddress(ctx, clientManager, types.Chain_eth, nil, []string{tmp_toaddress,}, done_watchaddress)

	token := "ZToken"
	go testSendTokenTx(ctx, clientManager, tmp_account.PrivateKey, tmp_toaddress, types.Chain_eth,
		&token, 10, done_sendTx)


	select {
	case <-done_sendTx:{
		L4g.Trace("SendTransaction done!")
	}
	case <-done_watchaddress:{
		L4g.Trace("Watch Address done!")
	}
	}

	clientManager.Close()

	L4g.Trace("exit main!")

	time.Sleep(1 * time.Second)
}

func testWatchAddress(ctx context.Context, clientManager *ClientManager,coin string, token *string, addresslist []string, done chan bool) {
	rcTxChannel := make(types.RechargeTxChannel)
	subscribe := clientManager.SubscribeTxRecharge(rcTxChannel)

	rcaCmd := NewRechargeAddressCmd("message id", types.Chain_eth, addresslist)
	clientManager.InsertRechargeAddress(rcaCmd)

	watch_address_channel := make(chan bool)

	subCtx, cancel := context.WithCancel(ctx)
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		defer subscribe.Unsubscribe()
		defer close(channel)

		exit := false
		for !exit {
			select {
			case rct := <-channel:{
				if rct==nil {
					L4g.Trace("Watch Address channel is close!")
				} else {
					L4g.Trace("Recharge Transaction : cointype:%s, information:%s.", rct.Coin_name, rct.Tx.String())
					if rct.Tx.State==types.Tx_state_confirmed || rct.Tx.State==types.Tx_state_unconfirmed {
						watch_address_channel <- true
					}
				}
			}
			case <-ctx.Done():{
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

func testSendTokenTx(ctx context.Context, clientManager *ClientManager, privatekey, to, coin string,
	token *string, value uint64, done chan bool) {
	txCmd := NewSendTxCmd("message id", coin, privatekey, to, token, value)
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
			case cmdTx := <-txStateChannel:{
				if cmdTx==nil {
					L4g.Trace("Transaction Command Channel is closed!")
					txok_channel <- false
				} else {
					fmt.Printf("Transaction state changed, transaction information:%s\n",
						cmdTx.Tx.String())

					if cmdTx.Tx.State == types.Tx_state_confirmed {
						L4g.Trace("Transaction is confirmed! success!!!")
						txok_channel <- true
					}

					if cmdTx.Tx.State == types.Tx_state_unconfirmed {
						L4g.Trace("Transaction is unconfirmed! failed!!!!")
						txok_channel <- false
					}
				}
			}
			case <-ctx.Done():{
				close = true
			}
			}
		}
	}(subCtx, txStateChannel)

	select {
	case <-txok_channel: {
		cancel()
	}
	}

	done <- true

}

