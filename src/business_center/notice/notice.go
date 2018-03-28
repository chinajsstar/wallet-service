package notice

import (
	"blockchain_server/types"
	"context"
	"fmt"
	"sync"
)

type Notice struct {
	rechargeChannel types.RechargeTxChannel
	cmdTxChannel    types.CmdTxChannel
	waitGroup       sync.WaitGroup
	ctx             context.Context
}

func NewNotice(ctx context.Context, rechargeChannel types.RechargeTxChannel, cmdTxChannel types.CmdTxChannel) *Notice {
	instance := new(Notice)
	instance.rechargeChannel = rechargeChannel
	instance.cmdTxChannel = cmdTxChannel
	instance.ctx = ctx
	return instance
}

func (ntc *Notice) Start() {
	ntc.recvRechargeTxChannel()
	ntc.recvCmdTxChannel()
}

func (ntc *Notice) Stop() {
	ntc.waitGroup.Wait()
}

func (ntc *Notice) recvRechargeTxChannel() {
	ntc.waitGroup.Add(1)
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		for {
			select {
			case rct := <-channel:
				{
					fmt.Printf("Recharge Transaction : cointype:%s, information:%s.", rct.Coin_name, rct.Tx.String())
				}
			case <-ctx.Done():
				{
					fmt.Println("RechangeTx context done, because : ", ctx.Err())
					ntc.waitGroup.Done()
					return
				}
			}
		}
	}(ntc.ctx, ntc.rechargeChannel)
}

func (ntc *Notice) recvCmdTxChannel() {
	ntc.waitGroup.Add(1)
	go func(ctx context.Context, channel types.CmdTxChannel) {
		for {
			select {
			case cmdTx := <-channel:
				{
					fmt.Printf("Transaction state changed, transaction information:%s\n",
						cmdTx.Tx.String())

					if cmdTx.Tx.State == types.Tx_state_confirmed {
						fmt.Println("Transaction is confirmed! success!!!")
					}

					if cmdTx.Tx.State == types.Tx_state_unconfirmed {
						fmt.Println("Transaction is unconfirmed! failed!!!!")
					}
				}
			case <-ctx.Done():
				fmt.Println("TxState context done, because : ", ctx.Err())
				ntc.waitGroup.Done()
				return
			}
		}
	}(ntc.ctx, ntc.cmdTxChannel)
}
