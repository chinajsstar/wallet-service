package service

import (
	"testing"
	"blockchain_server/types"
	"blockchain_server/chains/eth"
	"fmt"
	"context"
)

func TestNetCmdSvr(t *testing.T) {
	cmdSvr := &NetcmdHandler{}
	client, err := eth.NewClient()
	rechTxChannal := make(types.RechargeTxChannel)
	if nil!=err {
		fmt.Printf("create client:%s error:%s", types.Chain_eth, err.Error() )
		return
	}

	// 设置需要关注的地址
	// cmdSvr.SubscribeInComeTxWithAddresses()
	// 启动eth的服务!!
	// 第三个参数, 是用来接收充值的消息的
	cmdSvr.StartClient(types.Chain_eth, client, rechTxChannal)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()


	// 监控充值的交易
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		exit := false
		for !exit {
			select {
			case rct := <-channel:{
				fmt.Printf("Recharge Transaction : cointype:%s, information:%s.", rct.Coin_name, rct.Tx.String())
			}
			case <-ctx.Done():{
				fmt.Println("context done, because : ", ctx.Err())
				exit = true
			}
			}
		}
	}(ctx, rechTxChannal)

	// 创建并发送Transaction, 订阅只需要调用一次, 所有的Send的交易都会通过这个订阅channel传出来
	// cmdSvr.SubscribeTxStateChange()
	// newTransferCmd()
	// cmdSvr.SendTx()
}
