package address

import (
	"blockchain_server/service"
	"blockchain_server/types"
	. "business_center/def"
	"business_center/mysqlpool"
	"business_center/redispool"
	"context"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"log"
	"math"
	"time"
)

func (a *Address) generateAddress(userProperty *UserProperty, assetProperty *AssetProperty, count int) []UserAddress {
	cmd := service.NewAccountCmd("", assetProperty.AssetName, 1)
	userAddress := make([]UserAddress, 0)
	for i := 0; i < count; i++ {
		accounts, err := a.wallet.NewAccounts(cmd)
		if err != nil {
			CheckError(ErrorWallet, err.Error())
			return []UserAddress{}
		}
		nowTM := time.Now().Unix()
		data := UserAddress{
			UserKey:         userProperty.UserKey,
			UserClass:       userProperty.UserClass,
			AssetID:         assetProperty.AssetID,
			Address:         accounts[0].Address,
			PrivateKey:      accounts[0].PrivateKey,
			AvailableAmount: 0,
			FrozenAmount:    0,
			Enabled:         1,
			CreateTime:      nowTM,
			UpdateTime:      nowTM,
		}

		//添加地址监控
		cmd := service.NewRechargeAddressCmd("", assetProperty.AssetName, []string{data.Address})
		err = a.wallet.InsertRechargeAddress(cmd)
		if err != nil {
			CheckError(ErrorWallet, err.Error())
			return []UserAddress{}
		}
		userAddress = append(userAddress, data)
	}
	err := mysqlpool.AddUserAddress(userAddress)
	if err != nil {
		return []UserAddress{}
	}
	err = mysqlpool.AddUserAccount(userProperty.UserKey, userProperty.UserClass, assetProperty.AssetID)
	if err != nil {
		return []UserAddress{}
	}
	return userAddress
}

func (a *Address) recvRechargeTxChannel() {
	a.waitGroup.Add(1)
	go func(ctx context.Context, channel types.RechargeTxChannel) {
		c := redispool.Get()
		defer c.Close()

		for {
			select {
			case rct := <-channel:
				{
					assetProperty, ok := mysqlpool.QueryAssetPropertyByName(rct.Coin_name)
					if !ok {
						continue
					}

					blockin2 := TransactionBlockin2{
						AssetID:       assetProperty.AssetID,
						Hash:          rct.Tx.Tx_hash,
						MinerFee:      int64(rct.Tx.Minerfee()),
						BlockinHeight: int64(rct.Tx.InBlock),
						OrderID:       "",
						Time:          int64(rct.Tx.Time),
					}

					switch rct.Tx.State {
					case types.Tx_state_mined: //入块
						{
							var blockin TransactionBlockin
							blockin.AssetID = assetProperty.AssetID
							blockin.AssetName = assetProperty.AssetName
							blockin.Hash = rct.Tx.Tx_hash
							blockin.Status = 0
							blockin.MinerFee = int64(rct.Tx.Minerfee())
							blockin.BlockinHeight = int64(rct.Tx.InBlock)
							blockin.BlockinTime = int64(rct.Tx.Time)
							blockin.OrderID = ""

							a.transactionBegin(&blockin, rct.Tx)

							blockin2.Status = 0
						}
					case types.Tx_state_confirmed: //确认
						{
							var status TransactionStatus
							status.AssetID = assetProperty.AssetID
							status.AssetName = assetProperty.AssetName
							status.Hash = rct.Tx.Tx_hash
							status.Status = 1
							status.ConfirmHeight = int64(rct.Tx.ConfirmatedHeight)
							status.ConfirmTime = int64(rct.Tx.Time)
							status.UpdateTime = time.Now().Unix()
							status.OrderID = ""

							a.transactionFinish(&status, rct.Tx)

							blockin2.Status = 1
						}
					case types.Tx_state_unconfirmed: //失败
						{
							var status TransactionStatus
							status.AssetID = assetProperty.AssetID
							status.AssetName = assetProperty.AssetName
							status.Hash = rct.Tx.Tx_hash
							status.Status = 2
							status.ConfirmHeight = int64(rct.Tx.ConfirmatedHeight)
							status.ConfirmTime = int64(rct.Tx.Time)
							status.UpdateTime = time.Now().Unix()
							status.OrderID = ""

							a.transactionFinish(&status, rct.Tx)

							blockin2.Status = 2
						}
					default:
						continue
					}

					if blockin2.Status == 0 {
						a.transactionBegin2(&blockin2, rct.Tx)
					} else {
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
					assetProperty, ok := mysqlpool.QueryAssetPropertyByName(cmdTx.Coinname)
					if !ok {
						continue
					}

					switch cmdTx.Tx.State {
					case types.Tx_state_commited:
					case types.Tx_state_mined: //入块
						{
							var blockin TransactionBlockin
							blockin.AssetID = assetProperty.AssetID
							blockin.AssetName = assetProperty.AssetName
							blockin.Hash = cmdTx.Tx.Tx_hash
							blockin.Status = 0
							blockin.MinerFee = int64(cmdTx.Tx.Minerfee())
							blockin.BlockinHeight = int64(cmdTx.Tx.InBlock)
							blockin.BlockinTime = int64(cmdTx.Tx.Time)
							blockin.OrderID = cmdTx.NetCmd.MsgId

							a.transactionBegin(&blockin, cmdTx.Tx)
						}
					case types.Tx_state_confirmed: //确认
						{
							var status TransactionStatus
							status.AssetID = assetProperty.AssetID
							status.AssetName = assetProperty.AssetName
							status.Hash = cmdTx.Tx.Tx_hash
							status.Status = 1
							status.ConfirmHeight = int64(cmdTx.Tx.ConfirmatedHeight)
							status.ConfirmTime = int64(cmdTx.Tx.Time)
							status.OrderID = cmdTx.NetCmd.MsgId
							status.UpdateTime = time.Now().Unix()

							a.transactionFinish(&status, cmdTx.Tx)
						}
					case types.Tx_state_unconfirmed: //失败
						{
							var status TransactionStatus
							status.AssetID = assetProperty.AssetID
							status.AssetName = assetProperty.AssetName
							status.Hash = cmdTx.Tx.Tx_hash
							status.Status = 2
							status.ConfirmHeight = int64(cmdTx.Tx.ConfirmatedHeight)
							status.ConfirmTime = int64(cmdTx.Tx.Time)
							status.OrderID = cmdTx.NetCmd.MsgId
							status.UpdateTime = time.Now().Unix()

							a.transactionFinish(&status, cmdTx.Tx)
						}
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

func (a *Address) transactionBegin(blockin *TransactionBlockin, transfer *types.Transfer) error {
	db := mysqlpool.Get()

	if len(blockin.OrderID) > 0 {
		row := db.QueryRow("select user_key, asset_id, address, amount, wallet_fee, hash from withdrawal_order"+
			" where order_id = ?;",
			blockin.OrderID)

		var tn TransactionNotic
		row.Scan(&tn.UserKey, &tn.AssetID, &tn.Address, &tn.Amount, &tn.WalletFee, &tn.Hash)

		if len(tn.Hash) <= 0 {
			db.Exec("update withdrawal_order set hash = ? where order_id = ?;", blockin.Hash, blockin.OrderID)
		}

		tn.MsgID = 0
		tn.Type = TypeWithdrawal
		tn.Status = StatusBlockin
		tn.BlockinHeight = blockin.BlockinHeight
		tn.Hash = blockin.Hash
		tn.Time = blockin.BlockinTime

		a.sendTransactionNotic(&tn)
	}

	_, err := db.Exec("insert transaction_blockin (asset_id, hash, status, miner_fee, blockin_height, blockin_time,"+
		" confirm_height, confirm_time, order_id) values (?, ?, ?, ?, ?, ?, ?, ?, ?);",
		blockin.AssetID, blockin.Hash, blockin.Status, blockin.MinerFee, blockin.BlockinHeight,
		time.Unix(blockin.BlockinTime, 0).UTC().Format(TimeFormat), blockin.BlockinHeight,
		time.Unix(blockin.BlockinTime, 0).UTC().Format(TimeFormat), blockin.OrderID)

	if err != nil {
		return err
	}

	return a.preSettlement(blockin, transfer)
}

func (a *Address) transactionBegin2(blockin *TransactionBlockin2, transfer *types.Transfer) error {
	return nil
}

func (a *Address) preSettlement(blockin *TransactionBlockin, transfer *types.Transfer) error {
	var detail TransactionDetail
	blockin.Detail = make([]TransactionDetail, 0)

	switch blockin.AssetName {
	case "btc":
		{

		}
	case "eth":
		{
			//from
			detail.AssetID = blockin.AssetID
			detail.Address = transfer.From
			detail.TransType = "from"
			detail.Amount = -int64(transfer.Value)
			detail.Hash = blockin.Hash
			detail.DetailID = a.generateUUID()
			blockin.Detail = append(blockin.Detail, detail)

			//to
			detail.AssetID = blockin.AssetID
			detail.Address = transfer.To
			detail.TransType = "to"
			detail.Amount = int64(transfer.Value)
			detail.Hash = blockin.Hash
			detail.DetailID = a.generateUUID()
			blockin.Detail = append(blockin.Detail, detail)

			//miner_fee
			detail.AssetID = blockin.AssetID
			detail.Address = transfer.From
			detail.TransType = "miner_fee"
			detail.Amount = -int64(transfer.Minerfee())
			detail.Hash = blockin.Hash
			detail.DetailID = a.generateUUID()
			blockin.Detail = append(blockin.Detail, detail)
		}
	default:
		return nil
	}

	Tx, err := mysqlpool.Get().Begin()
	if err != nil {
		return err
	}

	for _, detail := range blockin.Detail {
		userAddress, ok := mysqlpool.QueryUserAddressByIDAddress(blockin.AssetID, detail.Address)
		Tx.Exec("insert transaction_detail "+
			"(asset_id, address, trans_type, amount, hash, detail_id) "+
			"values (?, ?, ?, ?, ?, ?);",
			blockin.AssetID, detail.Address, detail.TransType,
			detail.Amount, detail.Hash, detail.DetailID)

		if ok {
			Tx.Exec("update user_address set available_amount = available_amount + ?, update_time = ?"+
				" where asset_id = ? and address = ?;",
				detail.Amount, time.Now().UTC().Format(TimeFormat), userAddress.AssetID, userAddress.Address)
		}

		switch detail.TransType {
		case "from":
		case "to":
			if ok && userAddress.UserClass == 0 {

				//充值入块消息处理
				var tn TransactionNotic
				tn.UserKey = userAddress.UserKey
				tn.MsgID = 0
				tn.Type = TypeDeposit
				tn.Status = StatusBlockin
				tn.BlockinHeight = blockin.BlockinHeight
				tn.AssetID = blockin.AssetID
				tn.Address = userAddress.Address
				tn.Amount = detail.Amount
				tn.WalletFee = 0
				tn.Hash = blockin.Hash
				tn.Time = blockin.BlockinTime

				a.sendTransactionNotic(&tn)
			}
		case "miner_fee":
		case "change":
		}
	}

	Tx.Commit()

	return nil
}

func (a *Address) transactionFinish(status *TransactionStatus, transfer *types.Transfer) error {
	db := mysqlpool.Get()

	var blockin TransactionBlockin
	err := a.preTransactionFinish(status, &blockin, transfer)
	if err != nil {
		return err
	}

	_, err = db.Exec("insert transaction_status (asset_id, hash, status, confirm_height, confirm_time, update_time, order_id) "+
		"values (?, ?, ?, ?, ?, ?, ?);",
		status.AssetID, status.Hash, status.Status, status.ConfirmHeight,
		time.Unix(status.ConfirmTime, 0).UTC().Format(TimeFormat),
		time.Unix(status.UpdateTime, 0).UTC().Format(TimeFormat),
		status.OrderID)

	if err != nil {
		return nil
	}

	db.Exec("update transaction_blockin set status = ?, confirm_height = ?, confirm_time = ?"+
		" where asset_id = ? and hash = ?;",
		status.Status, status.ConfirmHeight, time.Unix(status.ConfirmTime, 0).UTC().Format(TimeFormat),
		status.AssetID, status.Hash)

	rows, _ := db.Query("select asset_id, address, trans_type, amount, hash, detail_id from transaction_detail"+
		" where asset_id = ? and hash = ?;",
		status.AssetID, status.Hash)

	var detail TransactionDetail
	for rows.Next() {
		err := rows.Scan(&detail.AssetID, &detail.Address, &detail.TransType, &detail.Amount,
			&detail.Hash, &detail.DetailID)
		if err == nil {
			userAddress, ok := mysqlpool.QueryUserAddressByIDAddress(blockin.AssetID, detail.Address)
			switch detail.TransType {
			case "from":
			case "to":
				if ok && userAddress.UserClass == 0 {
					//充值帐户余额修改
					db.Exec("update user_account set available_amount = available_amount + ?,"+
						" update_time = ? where user_key = ? and asset_id = ?;",
						detail.Amount, time.Now().UTC().Format(TimeFormat), userAddress.UserKey, detail.AssetID)

					//充值确认消息处理
					var tn TransactionNotic
					tn.UserKey = userAddress.UserKey
					tn.MsgID = 0
					tn.Type = TypeDeposit
					tn.Status = StatusConfirm
					tn.BlockinHeight = blockin.BlockinHeight
					tn.AssetID = blockin.AssetID
					tn.Address = detail.Address
					tn.Amount = detail.Amount
					tn.WalletFee = 0
					tn.Hash = blockin.Hash
					tn.Time = blockin.BlockinTime

					a.sendTransactionNotic(&tn)
				}
			case "miner_fee":
			case "change":
			}
		}
	}

	//结算提币订单
	if len(blockin.OrderID) > 0 {
		row := db.QueryRow("select user_key, asset_id, address, amount, wallet_fee, hash from withdrawal_order"+
			" where order_id = ?;", blockin.OrderID)

		var tn TransactionNotic
		err := row.Scan(&tn.UserKey, &tn.AssetID, &tn.Address, &tn.Amount, &tn.WalletFee, &tn.Hash)
		if err == nil {
			tn.MsgID = 0
			tn.Type = TypeWithdrawal
			tn.Status = StatusConfirm
			tn.BlockinHeight = blockin.BlockinHeight
			tn.Time = blockin.BlockinTime

			a.sendTransactionNotic(&tn)

			_, err := db.Exec("update user_account set frozen_amount = frozen_amount - ?, update_time = ?"+
				" where user_key = ? and asset_id = ?;",
				tn.Amount+tn.WalletFee, time.Now().UTC().Format(TimeFormat), tn.UserKey, tn.AssetID)

			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}

	return nil
}

func (a *Address) preTransactionFinish(status *TransactionStatus, blockin *TransactionBlockin, transfer *types.Transfer) error {
	db := mysqlpool.Get()
	blockin.AssetName = status.AssetName
	row := db.QueryRow("select asset_id, hash, status, miner_fee, blockin_height, unix_timestamp(blockin_time), order_id"+
		" from transaction_blockin where asset_id = ? and hash = ?;",
		status.AssetID, status.Hash)

	err := row.Scan(&blockin.AssetID, &blockin.Hash, &blockin.Status, &blockin.MinerFee,
		&blockin.BlockinHeight, &blockin.BlockinTime, &blockin.OrderID)
	if err != nil {
		blockin.AssetID = status.AssetID
		blockin.AssetName = status.AssetName
		blockin.Hash = transfer.Tx_hash
		blockin.Status = 0
		blockin.MinerFee = int64(transfer.Minerfee())
		blockin.BlockinHeight = int64(transfer.InBlock)
		blockin.BlockinTime = int64(transfer.Time)
		blockin.OrderID = status.OrderID
		return a.transactionBegin(blockin, transfer)
	}
	return nil
}

func (a *Address) sendTransactionNotic(tn *TransactionNotic) error {
	db := mysqlpool.Get()

	ret, err := db.Exec("insert into transaction_notice (user_key, msg_id,"+
		" type, status, blockin_height, asset_id, address, amount, wallet_fee, hash, time)"+
		" select ?, count(*)+1, ?, ?, ?, ?, ?, ?, ?, ?, ? from transaction_notice where user_key = ?;",
		tn.UserKey, tn.Type, tn.Status, tn.BlockinHeight, tn.AssetID, tn.Address,
		tn.Amount, tn.WalletFee, tn.Hash, time.Unix(tn.Time, 0).Format(TimeFormat), tn.UserKey)
	if err != nil {
		return err
	}

	insertID, err := ret.LastInsertId()
	if err != nil {
		return err
	}

	row := db.QueryRow("select msg_id from transaction_notice where id = ?;", insertID)
	row.Scan(&tn.MsgID)

	return nil

	// push notify by liuheng
	b, err := json.Marshal(tn)
	if err != nil {
		log.Println("Push Error: json Marshal")
	} else {
		a.callback(tn.UserKey, string(b))
	}
	return nil
}

func (a *Address) generateUUID() string {
	uID := ""
	u, _ := uuid.NewV4()
	uID = fmt.Sprintf("0x%x", u.Bytes())
	return uID
}

func responsePagination(query string, totalLines int) map[string]interface{} {
	resMap := make(map[string]interface{})
	resMap["total_lines"] = totalLines

	if len(query) > 0 {
		var queryMap map[string]interface{}
		err := json.Unmarshal([]byte(query), &queryMap)
		if err != nil {
			return resMap
		}
		if value, ok := queryMap["page_index"]; ok {
			resMap["page_index"] = value
		}
		if value, ok := queryMap["max_disp_lines"]; ok {
			resMap["max_disp_lines"] = value
		}
	}
	return resMap
}

func packJson(v interface{}) string {
	s, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(s)
}

func unpackJson(s string) ParamsMapping {
	params := ParamsMapping{UserKey: "", UserOrderID: "", AssetID: 0, Address: "", Amount: 0, Count: 0}
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(s), &jsonMap)
	if err != nil {
		return params
	}

	for k, v := range jsonMap {
		switch k {
		case "user_key":
			if value, ok := v.(string); ok {
				params.UserKey = value
			}
		case "user_order_id":
			if value, ok := v.(string); ok {
				params.UserOrderID = value
			}
		case "asset_id":
			if value, ok := v.(float64); ok {
				params.AssetID = int(value)
			}
		case "address":
			if value, ok := v.(string); ok {
				params.Address = value
			}
		case "amount":
			if value, ok := v.(float64); ok {
				params.Amount = int64(value * math.Pow10(8))
			}
		case "count":
			if value, ok := v.(float64); ok {
				params.Count = int(value)
			}
		}
	}
	return params
}
