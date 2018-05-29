package monitor

import (
	"blockchain_server/service"
	"blockchain_server/types"
	. "business/def"
	"business/mysqlpool"
	"context"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"sync"
)

type Monitor struct {
	wallet          *service.ClientManager
	callback        PushMsgCallback
	rechargeChannel types.RechargeTxChannel
	cmdTxChannel    types.CmdTxChannel
	waitGroup       sync.WaitGroup
	ctx             context.Context
}

func (m *Monitor) Run(ctx context.Context, wallet *service.ClientManager, callback PushMsgCallback) {
	m.wallet = wallet
	m.callback = callback
	m.ctx = ctx

	m.rechargeChannel = make(types.RechargeTxChannel)
	m.cmdTxChannel = make(types.CmdTxChannel)

	m.recvRechargeTxChannel()
	m.recvCmdTxChannel()

	m.wallet.SubscribeTxRecharge(m.rechargeChannel)
	m.wallet.SubscribeTxCmdState(m.cmdTxChannel)

	//添加监控地址
	if userAddress, ok := mysqlpool.QueryUserAddress(nil); ok {
		for _, v := range userAddress {
			if assetProperty, ok := mysqlpool.QueryAssetPropertyByName(v.AssetName); ok {
				assetName := assetProperty.AssetName
				if assetProperty.IsToken > 0 {
					assetName = assetProperty.ParentName
				}
				rcaCmd := service.NewRechargeAddressCmd("", assetName, []string{v.Address})
				m.wallet.InsertRechargeAddress(rcaCmd)
			}
		}
	}
}

func (m *Monitor) Stop() {
	m.waitGroup.Wait()
}

func (m *Monitor) recvRechargeTxChannel() {
	m.waitGroup.Add(1)
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		for {
			select {
			case rct := <-channel:
				{
					if rct.Err != nil {
						l4g.Error("%s", rct.Err.Error())
						continue
					}
					assetProperty, ok := mysqlpool.QueryAssetPropertyByName(rct.Coin_name)
					if !ok {
						continue
					}

					blockin := TransactionBlockin{
						AssetName:     assetProperty.AssetName,
						Hash:          rct.Tx.Tx_hash,
						MinerFee:      rct.Tx.Fee,
						BlockinHeight: int64(rct.Tx.InBlock),
						OrderID:       "",
					}

					switch rct.Tx.State {
					case types.Tx_state_mined: //入块
						blockin.Status = StatusBlockin
					case types.Tx_state_confirmed: //确认
						blockin.Status = StatusConfirm
					case types.Tx_state_unconfirmed: //失败
						blockin.Status = StatusFail
					default:
						continue
					}

					if blockin.Status == StatusBlockin {
						Blockin(&blockin, rct.Tx, m.callback)
					} else if blockin.Status == StatusConfirm {
						Blockin(&blockin, rct.Tx, m.callback)
						Confirm(&blockin, rct.Tx, m.callback)
					}
				}
			case <-ctx.Done():
				{
					fmt.Println("RechangeTx context done, because : ", ctx.Err())
					m.waitGroup.Done()
					return
				}
			}
		}
	}(m.ctx, m.rechargeChannel)
}

func (m *Monitor) recvCmdTxChannel() {
	m.waitGroup.Add(1)
	go func(ctx context.Context, channel types.CmdTxChannel) {
		for {
			select {
			case cmdTx := <-channel:
				{
					if cmdTx.Error != nil {
						l4g.Error("%s", cmdTx.Error.Message)
						continue
					}

					assetProperty, ok := mysqlpool.QueryAssetPropertyByName(cmdTx.Coinname)
					if !ok {
						continue
					}

					blockin := TransactionBlockin{
						AssetName:     assetProperty.AssetName,
						Hash:          cmdTx.Tx.Tx_hash,
						MinerFee:      cmdTx.Tx.Fee,
						BlockinHeight: int64(cmdTx.Tx.InBlock),
						OrderID:       cmdTx.NetCmd.MsgId,
					}

					switch cmdTx.Tx.State {
					case types.Tx_state_mined: //入块
						blockin.Status = StatusBlockin
					case types.Tx_state_confirmed: //确认
						blockin.Status = StatusConfirm
					case types.Tx_state_unconfirmed: //失败
						blockin.Status = StatusFail
					default:
						continue
					}

					if blockin.Status == StatusBlockin {
						Blockin(&blockin, cmdTx.Tx, m.callback)
					} else if blockin.Status == StatusConfirm {
						Blockin(&blockin, cmdTx.Tx, m.callback)
						Confirm(&blockin, cmdTx.Tx, m.callback)
					}
				}
			case <-ctx.Done():
				fmt.Println("TxState context done, because : ", ctx.Err())
				m.waitGroup.Done()
				return
			}
		}
	}(m.ctx, m.cmdTxChannel)
}
