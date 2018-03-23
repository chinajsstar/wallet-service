package main

import (
	"fmt"
	"context"
	"blockchain_server/chains/eth"
	"blockchain_server/service"
	"blockchain_server/types"
	l4g "github.com/alecthomas/log4go"
	"github.com/ethereum/go-ethereum/crypto"
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
	clientManager.SetRechargeAddress(rcaCmd)

	/*********创建监控充币地址channael*********/
	rcTxChannel := make(types.RechargeTxChannel)
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		exit := false
		for !exit {
			select {
			case rct := <-channel:{
				fmt.Printf("Recharge Transaction : cointype:%s, information:%s.", rct.Coin_name, rct.Tx.String())
			}
			case <-ctx.Done():{
				fmt.Println("RechangeTx context done, because : ", ctx.Err())
				exit = true
			}
			}
		}
	}(ctx, rcTxChannel)


	/*********开启服务!!!!!*********/
	ctx3, _ := context.WithCancel(ctx)
	clientManager.Start(ctx3, rcTxChannel)



	/*********监控提币交易的channel*********/
	txStateChannel := make(types.TxStateChange_Channel)

	// 创建并发送Transaction, 订阅只需要调用一次, 所有的Send的交易都会通过这个订阅channel传回来
	/*********执行提币命令*********/
	l4g.Trace("SendTransaction from :%s to:%s", tmp_account.Address, tmp_toaddress)
	key, _ := eth.ParseChiperkey(tmp_account.PrivateKey)
	if key != nil {
		l4g.Trace("SendTransaction from :%s to:%s", crypto.PubkeyToAddress(key.PublicKey).String(), tmp_toaddress)
	}

	txCmd := types.NewTxCmd("message id of transaction command", types.Chain_eth, tmp_account.PrivateKey, tmp_toaddress, 1)
	clientManager.SendTx(txCmd)

	subTx := clientManager.SubscribeTxStateChange(txStateChannel)
	ctx2, _ := context.WithCancel(ctx)
	func(ctx2 context.Context, txstateChannel types.TxStateChange_Channel) {
		close := false
		for !close {
			select {
			case cmdTx := <-txStateChannel:{
				fmt.Printf("Transaction state changed, transaction information:%s\n",
					cmdTx.Tx.String())
				if cmdTx.Tx.State == types.Tx_state_confirmed &&
					cmdTx.Tx.Confirmationsnumber==(cmdTx.Tx.PresentBlocknumber-cmdTx.Tx.OnBlocknumber) {
					fmt.Printf("Transaction success done!!!!\n")
					close = true
				}
				if cmdTx.Tx.State == types.Tx_state_unconfirmed {
					fmt.Printf("Transaction failed!!!!\n")
					close = true
				}
			}
			case <-ctx.Done():{
				close = true
			}
			}
		}
	}(ctx2, txStateChannel)

	// 去掉订阅本次的Transaction
	subTx.Unsubscribe()
}

