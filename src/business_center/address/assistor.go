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
	l4g "github.com/alecthomas/log4go"
	"github.com/satori/go.uuid"
	"log"
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
	err = mysqlpool.AddUserAccount(userProperty.UserKey, userProperty.UserClass, assetProperty.AssetName)
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

					blockin := TransactionBlockin{
						AssetName:     assetProperty.AssetName,
						Hash:          rct.Tx.Tx_hash,
						MinerFee:      int64(rct.Tx.Fee),
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
						blockin.Time = int64(rct.Tx.Time)
						a.transactionBegin(&blockin, rct.Tx)
					} else {
						blockin.Time = time.Now().Unix()
						a.transactionFinish(&blockin, rct.Tx)
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

					blockin := TransactionBlockin{
						AssetName:     assetProperty.AssetName,
						Hash:          cmdTx.Tx.Tx_hash,
						MinerFee:      int64(cmdTx.Tx.Fee),
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
						blockin.Time = int64(cmdTx.Tx.Time)
						a.transactionBegin(&blockin, cmdTx.Tx)
					} else {
						blockin.Time = time.Now().Unix()
						a.transactionFinish(&blockin, cmdTx.Tx)
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
		row := db.QueryRow("select user_key, asset_name, address, amount, pay_fee, hash"+
			" from withdrawal_order where order_id = ?;", blockin.OrderID)

		transNotice := TransactionNotice{
			MsgID:         0,
			Type:          TypeWithdrawal,
			Status:        StatusBlockin,
			BlockinHeight: blockin.BlockinHeight,
			Hash:          blockin.Hash,
			OrderID:       blockin.OrderID,
			Time:          blockin.Time,
		}

		err := row.Scan(&transNotice.UserKey, &transNotice.AssetName, &transNotice.Address, &transNotice.Amount,
			&transNotice.PayFee, &transNotice.Hash)
		if err == nil {
			if len(transNotice.Hash) <= 0 {
				db.Exec("update withdrawal_order set hash = ? where order_id = ?;", blockin.Hash, blockin.OrderID)
			}
			a.sendTransactionNotic(&transNotice)
		}
	}

	_, err := db.Exec("insert transaction_blockin (asset_name, hash, status, miner_fee, blockin_height, blockin_time,"+
		" confirm_height, confirm_time, order_id) values (?, ?, ?, ?, ?, ?, ?, ?, ?);",
		blockin.AssetName, blockin.Hash, blockin.Status, blockin.MinerFee, blockin.BlockinHeight,
		time.Unix(blockin.Time, 0).UTC().Format(TimeFormat), blockin.BlockinHeight,
		time.Unix(blockin.Time, 0).UTC().Format(TimeFormat), blockin.OrderID)
	if err == nil {
		return a.preSettlement(blockin, transfer)
	}
	return err
}

func (a *Address) preSettlement(blockin *TransactionBlockin, transfer *types.Transfer) error {
	detail := TransactionDetail{
		AssetName: blockin.AssetName,
		Hash:      blockin.Hash,
	}
	switch blockin.AssetName {
	case "btc":
		{
			//from
			detail.Address = transfer.From
			detail.TransType = "from"
			detail.Amount = -int64(transfer.Value)
			detail.DetailID = a.generateUUID()
			blockin.Detail = append(blockin.Detail, detail)

			//to
			detail.Address = transfer.To
			detail.TransType = "to"
			detail.Amount = int64(transfer.Value)
			detail.DetailID = a.generateUUID()
			blockin.Detail = append(blockin.Detail, detail)

			//miner_fee
			detail.Address = transfer.From
			detail.TransType = "miner_fee"
			detail.Amount = -int64(transfer.Fee)
			detail.DetailID = a.generateUUID()
			blockin.Detail = append(blockin.Detail, detail)
		}
	case "eth":
		{
			//from
			detail.Address = transfer.From
			detail.TransType = "from"
			detail.Amount = -int64(transfer.Value)
			detail.DetailID = a.generateUUID()
			blockin.Detail = append(blockin.Detail, detail)

			//to
			detail.Address = transfer.To
			detail.TransType = "to"
			detail.Amount = int64(transfer.Value)
			detail.DetailID = a.generateUUID()
			blockin.Detail = append(blockin.Detail, detail)

			//miner_fee
			detail.Address = transfer.From
			detail.TransType = "miner_fee"
			detail.Amount = -int64(transfer.Fee)
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

	transNotice := TransactionNotice{
		MsgID:     0,
		PayFee:    0,
		AssetName: blockin.AssetName,
		Hash:      blockin.Hash,
		Time:      blockin.Time,
		Status:    StatusBlockin,
	}

	for _, detail := range blockin.Detail {
		userAddress, ok := mysqlpool.QueryUserAddressByNameAddress(blockin.AssetName, detail.Address)
		Tx.Exec("insert transaction_detail "+
			"(asset_name, address, trans_type, amount, hash, detail_id) "+
			"values (?, ?, ?, ?, ?, ?);",
			blockin.AssetName, detail.Address, detail.TransType,
			detail.Amount, detail.Hash, detail.DetailID)

		if ok {
			Tx.Exec("update user_address set available_amount = available_amount + ?, update_time = ?"+
				" where asset_name = ? and address = ?;",
				detail.Amount, time.Now().UTC().Format(TimeFormat), userAddress.AssetName, userAddress.Address)
		}

		switch detail.TransType {
		case "from":
		case "to":
			if ok && userAddress.UserClass == 0 {

				//充值入块消息处理
				transNotice.UserKey = userAddress.UserKey
				transNotice.Type = TypeDeposit
				transNotice.BlockinHeight = blockin.BlockinHeight
				transNotice.Address = userAddress.Address
				transNotice.Amount = detail.Amount
				transNotice.OrderID = detail.DetailID

				a.sendTransactionNotic(&transNotice)
			}
		case "miner_fee":
		case "change":
		}
	}
	Tx.Commit()
	return nil
}

func (a *Address) transactionFinish(blockin *TransactionBlockin, transfer *types.Transfer) error {
	err := a.preTransactionFinish(*blockin, transfer)
	if err != nil {
		return err
	}

	db := mysqlpool.Get()
	_, err = db.Exec("insert transaction_status (asset_name, hash, status, confirm_height, confirm_time, update_time, order_id) "+
		"values (?, ?, ?, ?, ?, ?, ?);",
		blockin.AssetName, blockin.Hash, blockin.Status, transfer.ConfirmatedHeight,
		time.Unix(blockin.Time, 0).UTC().Format(TimeFormat),
		time.Unix(blockin.Time, 0).UTC().Format(TimeFormat),
		blockin.OrderID)
	if err != nil {
		return nil
	}

	db.Exec("update transaction_blockin set status = ?, confirm_height = ?, confirm_time = ?"+
		" where asset_name = ? and hash = ?;",
		blockin.Status, transfer.ConfirmatedHeight, time.Unix(blockin.Time, 0).UTC().Format(TimeFormat),
		blockin.AssetName, blockin.Hash)

	rows, _ := db.Query("select asset_name,address,trans_type,amount,hash,detail_id from transaction_detail"+
		" where asset_name = ? and hash = ?;",
		blockin.AssetName, blockin.Hash)

	var detail TransactionDetail
	for rows.Next() {
		err := rows.Scan(&detail.AssetName, &detail.Address, &detail.TransType, &detail.Amount,
			&detail.Hash, &detail.DetailID)
		if err == nil {
			userAddress, ok := mysqlpool.QueryUserAddressByNameAddress(blockin.AssetName, detail.Address)
			switch detail.TransType {
			case "from":
			case "to":
				if ok && userAddress.UserClass == 0 {
					//充值帐户余额修改
					db.Exec("update user_account set available_amount = available_amount + ?,"+
						" update_time = ? where user_key = ? and asset_name = ?;",
						detail.Amount, time.Now().UTC().Format(TimeFormat), userAddress.UserKey, detail.AssetName)

					//充值确认消息处理
					transNotice := TransactionNotice{
						UserKey:       userAddress.UserKey,
						MsgID:         0,
						Type:          TypeDeposit,
						Status:        StatusConfirm,
						BlockinHeight: blockin.BlockinHeight,
						AssetName:     blockin.AssetName,
						Address:       detail.Address,
						Amount:        detail.Amount,
						PayFee:        0,
						Hash:          blockin.Hash,
						OrderID:       detail.DetailID,
						Time:          blockin.Time,
					}
					a.sendTransactionNotic(&transNotice)
				}
			case "miner_fee":
			case "change":
			}
		}
	}

	//结算提币订单
	if len(blockin.OrderID) > 0 {
		row := db.QueryRow("select user_key, asset_name, address, amount, pay_fee, hash from withdrawal_order"+
			" where order_id = ?;", blockin.OrderID)

		transNotice := TransactionNotice{
			MsgID:         0,
			Type:          TypeWithdrawal,
			Status:        StatusConfirm,
			BlockinHeight: blockin.BlockinHeight,
			OrderID:       blockin.OrderID,
			Time:          blockin.Time,
		}
		err := row.Scan(&transNotice.UserKey, &transNotice.AssetName, &transNotice.Address, &transNotice.Amount, &transNotice.PayFee, &transNotice.Hash)
		if err == nil {
			a.sendTransactionNotic(&transNotice)
			_, err := db.Exec("update user_account set frozen_amount = frozen_amount - ?, update_time = ?"+
				" where user_key = ? and asset_name = ?;",
				transNotice.Amount+transNotice.PayFee, time.Now().UTC().Format(TimeFormat), transNotice.UserKey, transNotice.AssetName)
			if err != nil {
				l4g.Error(err.Error())
			}
		}
	}
	return nil
}

func (a *Address) preTransactionFinish(blockin TransactionBlockin, transfer *types.Transfer) error {
	db := mysqlpool.Get()
	count := 0
	row := db.QueryRow("select count(*) from transaction_blockin where asset_name = ? and hash = ?;",
		blockin.AssetName, blockin.Hash)
	row.Scan(&count)
	if count <= 0 {
		blockin.Status = 0
		blockin.Time = int64(transfer.Time)
		return a.transactionBegin(&blockin, transfer)
	}
	return nil
}

func (a *Address) sendTransactionNotic(t *TransactionNotice) error {
	db := mysqlpool.Get()
	ret, err := db.Exec("insert into transaction_notice (user_key, msg_id,"+
		" trans_type, status, blockin_height, asset_name, address, amount, pay_fee, hash, order_id, time)"+
		" select ?, count(*)+1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? from transaction_notice where user_key = ?;",
		t.UserKey, t.Type, t.Status, t.BlockinHeight, t.AssetName, t.Address,
		t.Amount, t.PayFee, t.Hash, t.OrderID, time.Unix(t.Time, 0).Format(TimeFormat), t.UserKey)
	if err != nil {
		return err
	}

	insertID, err := ret.LastInsertId()
	if err != nil {
		return err
	}

	row := db.QueryRow("select msg_id from transaction_notice where id = ?;", insertID)
	row.Scan(&t.MsgID)

	return nil

	// push notify by liuheng
	b, err := json.Marshal(t)
	if err != nil {
		log.Println("Push Error: json Marshal")
	} else {
		a.callback(t.UserKey, string(b))
	}
	return nil
}

func (a *Address) generateUUID() string {
	uID := ""
	u, _ := uuid.NewV4()
	uID = fmt.Sprintf("0x%x", u.Bytes())
	return uID
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
