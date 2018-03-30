package main

import (
	"fmt"
	"context"
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	l4g "github.com/alecthomas/log4go"
	"github.com/ethereum/go-ethereum/crypto"
	"time"
)


var (
	tmp_account = &types.Account{
		"0x04e2b6c9bfeacd4880d99790a03a3db4ad8d87c82bb7d72711b277a9a03e49743077f3ae6d0d40e6bc04eceba67c2b3ec670b22b30d57f9d6c42779a05fba097536c412af73be02d1642aecea9fa7082db301e41d1c3c2686a6a21ca431e7e8605f761d8e12d61ca77605b31d707abc3f17bc4a28f4939f352f283a48ed77fc274b039590cc2c43ef739bd3ea13e491316",
		"0x54B2E44D40D3Df64e38487DD4e145b3e6Ae25927"}
	tmp_toaddress = "0x498d8306dd26ab45d8b7dd4f07a40d2c744f54bc"
)

func init() {

}

func main() {
	clientManager := service.NewClientManager()
	client, err := eth.NewClient()
	if nil!=err {
		fmt.Printf("create client:%s error:%s", types.Chain_eth, err.Error() )
		return
	}

	txRechChannel := make(types.RechargeTxChannel, 56)
	subTxRech := clientManager.SubscribeTxRecharge(txRechChannel)
	txStateChannel := make(types.CmdTxChannel, 56)
	subTx := clientManager.SubscribeTxCmdState(txStateChannel)

	// add client instance to manager
	clientManager.AddClient(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/*********批量创建账号示例*********/
	accCmd := types.NewAccountCmd("message id of new account command message id", types.Chain_eth, 2)
	var accs []*types.Account
	accs, err = clientManager.NewAccounts(accCmd)
	for i, account := range accs {
		fmt.Printf("account[%d], crypt private key:%s, address:%s\n",
			i, account.PrivateKey, account.Address)
	}

	/*********添加监控地址示例*********/
	rcaCmd := types.NewRechargeAddressCmd("message id of monitor recharge address command", types.Chain_eth,
		[]string{accs[0].Address, tmp_toaddress})

	if err := clientManager.InsertRechargeAddress(rcaCmd); err!=nil {
		l4g.Error("insert recharge address  error:%s", err.Error())
	}

	/*********创建监控充币地址channael*********/
	watch_address_channel := make(chan bool)
	// 开启监控goroutine
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		exit := false
		for !exit {
			select {
			case rct := <-channel:{
				l4g.Trace("Recharge Transaction : cointype:%s, information:%s.", rct.Coin_name, rct.Tx.String())

				if rct.Tx.State == types.Tx_state_unconfirmed || rct.Tx.State==types.Tx_state_confirmed || rct.Err!=nil {
					watch_address_channel <- true
					if rct.Err!=nil {
						l4g.Error("Recharge Transaction error message:%s", rct.Err.Error())
					}
				}
			}
			case <-ctx.Done():{
				l4g.Trace("RechangeTx context done, because : ", ctx.Err())
				watch_address_channel <- true
				exit = true
			}
			}
		}
	}(ctx, txRechChannel)

	/*********监控提币交易的channel*********/
	txok_channel := make(chan bool)
	// 开启交易监控goroutine
	go func(tctx context.Context, xstateChannel types.CmdTxChannel) {
		close := false
		for !close {
			select {
			case cmdTx := <-txStateChannel:{
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
			case <-ctx.Done():{
				close = true
			}
			}
		}
	}(ctx, txStateChannel)

	/*********开启服务!!!!!*********/
	clientManager.Start()


	// 创建并发送Transaction, 订阅只需要调用一次, 所有的Send的交易都会通过这个订阅channel传回来
	/*********执行提币命令*********/
	l4g.Trace("SendTransaction from :%s to:%s", tmp_account.Address, tmp_toaddress)
	key, _ := eth.ParseChiperkey(tmp_account.PrivateKey)
	if key != nil {
		l4g.Trace("SendTransaction from :%s to:%s", crypto.PubkeyToAddress(key.PublicKey).String(), tmp_toaddress)
	}

	if false {
		txCmd := types.NewTxCmd("message id of transaction command", types.Chain_eth, tmp_account.PrivateKey, tmp_toaddress, 1)
		clientManager.SendTx(txCmd)
	}

	if ok := <-txok_channel; ok ||!ok {
		l4g.Trace("transaction gorouine already exited!")
	}

	if ok := <-watch_address_channel; ok ||!ok {
		l4g.Trace("watching address gorouine already exited!")
	}

	// 关闭所有client
	clientManager.Close()
	// 关闭订阅
	subTx.Unsubscribe()
	subTxRech.Unsubscribe()

	// 处罚ctx.done(), 结束地址监控和交易监控的goroutine
	cancel()

	l4g.Trace("exit main!")

	time.Sleep(1 * time.Second)
}

