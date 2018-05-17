package address

import (
	"blockchain_server/service"
	"blockchain_server/types"
	. "business_center/def"
	"business_center/mysqlpool"
	"business_center/transaction"
	"context"
	"encoding/json"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"time"
)

func (a *Address) generateAddress(userProperty *UserProperty, assetProperty *AssetProperty, count int) []UserAddress {
	assetName := assetProperty.AssetName
	if assetProperty.IsToken > 0 {
		assetName = assetProperty.ParentName
	}
	cmd := service.NewAccountCmd("", assetName, 1)
	userAddress := make([]UserAddress, 0)
	for i := 0; i < count; i++ {
		accounts, err := a.wallet.NewAccounts(cmd)
		if err != nil {
			CheckError(ErrorFailed, err.Error())
			return []UserAddress{}
		}
		nowTM := time.Now().Unix()
		data := UserAddress{
			UserKey:         userProperty.UserKey,
			UserClass:       userProperty.UserClass,
			AssetName:       assetProperty.AssetName,
			Address:         accounts[0].Address,
			PrivateKey:      accounts[0].PrivateKey,
			AvailableAmount: 0,
			FrozenAmount:    0,
			Enabled:         1,
			CreateTime:      nowTM,
			AllocationTime:  nowTM,
			UpdateTime:      nowTM,
		}

		//添加地址监控
		cmd := service.NewRechargeAddressCmd("", assetName, []string{data.Address})
		err = a.wallet.InsertRechargeAddress(cmd)
		if err != nil {
			CheckError(ErrorFailed, err.Error())
			return []UserAddress{}
		}
		userAddress = append(userAddress, data)
	}
	err := mysqlpool.AddUserAddress(userAddress)
	if err != nil {
		return []UserAddress{}
	}

	if userProperty.UserClass == 0 {
		err = mysqlpool.AddUserAccount(userProperty.UserKey, userProperty.UserClass, assetProperty.AssetName)
		if err != nil {
			return []UserAddress{}
		}
	}

	if userProperty.UserClass == 1 && assetProperty.IsToken == 0 {
		mysqlpool.CreateTokenAddress(userAddress)
	}

	return userAddress
}

func (a *Address) recvRechargeTxChannel() {
	a.waitGroup.Add(1)
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
						MinerFee:      float64(rct.Tx.Fee),
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
						transaction.Blockin(&blockin, rct.Tx, a.callback)
					} else if blockin.Status == StatusConfirm {
						transaction.Blockin(&blockin, rct.Tx, a.callback)
						transaction.Confirm(&blockin, rct.Tx, a.callback)
					}
				}
			case <-ctx.Done():
				{
					fmt.Println("RechangeTx context done, because : ", ctx.Err())
					a.waitGroup.Done()
					return
				}
			}
		}
	}(a.ctx, a.rechargeChannel)
}

func (a *Address) recvCmdTxChannel() {
	a.waitGroup.Add(1)
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
						MinerFee:      float64(cmdTx.Tx.Fee),
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
						transaction.Blockin(&blockin, cmdTx.Tx, a.callback)
					} else if blockin.Status == StatusConfirm {
						transaction.Blockin(&blockin, cmdTx.Tx, a.callback)
						transaction.Confirm(&blockin, cmdTx.Tx, a.callback)
					}
				}
			case <-ctx.Done():
				fmt.Println("TxState context done, because : ", ctx.Err())
				a.waitGroup.Done()
				return
			}
		}
	}(a.ctx, a.cmdTxChannel)
}

func responsePagination(queryMap map[string]interface{}, totalLines int) map[string]interface{} {
	resMap := make(map[string]interface{})
	resMap["total_lines"] = totalLines

	if len(queryMap) > 0 {
		if value, ok := queryMap["page_index"]; ok {
			resMap["page_index"] = value
		}
		if value, ok := queryMap["max_disp_lines"]; ok {
			resMap["max_disp_lines"] = value
		}
	}
	return resMap
}

func responseJson(v interface{}) string {
	s, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(s)
}
