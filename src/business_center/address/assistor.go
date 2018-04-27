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
	"time"
)

func (a *Address) generateAddress(userID string, userClass int,
	assetID int, assetName string, count int) []UserAddress {
	userAddresses := make([]UserAddress, 0)
	cmd := service.NewAccountCmd("", assetName, 1)

	for i := 0; i < count; i++ {
		accounts, err := a.wallet.NewAccounts(cmd)
		if err != nil {
			fmt.Printf("generateAddress NewAccounts Error : %s\n", err.Error())
			return userAddresses
		}
		var userAddress UserAddress
		userAddress.UserKey = userID
		userAddress.UserClass = userClass
		userAddress.AssetID = assetID
		userAddress.AssetName = assetName
		userAddress.Address = accounts[0].Address
		userAddress.PrivateKey = accounts[0].PrivateKey
		userAddress.AvailableAmount = 0
		userAddress.FrozenAmount = 0
		userAddress.Enabled = 1
		userAddress.CreateTime = time.Now().Unix()
		userAddress.UpdateTime = time.Now().Unix()

		userAddresses = append(userAddresses, userAddress)
	}
	return userAddresses
}

func (a *Address) addUserAddress(userAddress []UserAddress) []string {
	var addresses []string
	tx, err := mysqlpool.Get().Begin()
	if err != nil {
		return addresses
	}

	for _, v := range userAddress {
		_, err := tx.Exec("insert user_address (user_key, user_class, asset_id, address, private_key,"+
			" available_amount, frozen_amount, enabled, create_time, update_time) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);",
			v.UserKey, v.UserClass, v.AssetID, v.Address, v.PrivateKey, v.AvailableAmount, v.FrozenAmount, v.Enabled,
			time.Unix(v.CreateTime, 0).UTC().Format(TimeFormat),
			time.Unix(v.UpdateTime, 0).UTC().Format(TimeFormat))
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		addresses = append(addresses, v.Address)
	}
	tx.Commit()

	return addresses
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
					assetProperty, ok := mysqlpool.QueryAllAssetProperty()[rct.Coin_name]
					if !ok {
						continue
					}

					switch rct.Tx.State {
					case types.Tx_state_mined: //入块
						{
							var blockin TransactionBlockin
							blockin.AssetID = assetProperty.ID
							blockin.AssetName = assetProperty.Name
							blockin.Hash = rct.Tx.Tx_hash
							blockin.Status = 0
							blockin.MinerFee = int64(rct.Tx.Minerfee())
							blockin.BlockinHeight = int64(rct.Tx.InBlock)
							blockin.BlockinTime = int64(rct.Tx.Time)
							blockin.OrderID = ""

							a.transactionBegin(&blockin, rct.Tx)
						}
					case types.Tx_state_confirmed: //确认
						{
							var status TransactionStatus
							status.AssetID = assetProperty.ID
							status.AssetName = assetProperty.Name
							status.Hash = rct.Tx.Tx_hash
							status.Status = 1
							status.ConfirmHeight = int64(rct.Tx.ConfirmatedHeight)
							status.ConfirmTime = int64(rct.Tx.Time)
							status.UpdateTime = time.Now().Unix()
							status.OrderID = ""

							a.transactionFinish(&status, rct.Tx)
						}
					case types.Tx_state_unconfirmed: //失败
						{
							var status TransactionStatus
							status.AssetID = assetProperty.ID
							status.AssetName = assetProperty.Name
							status.Hash = rct.Tx.Tx_hash
							status.Status = 2
							status.ConfirmHeight = int64(rct.Tx.ConfirmatedHeight)
							status.ConfirmTime = int64(rct.Tx.Time)
							status.UpdateTime = time.Now().Unix()
							status.OrderID = ""

							a.transactionFinish(&status, rct.Tx)
						}
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
					assetProperty, ok := mysqlpool.QueryAllAssetProperty()[cmdTx.Coinname]
					if !ok {
						continue
					}

					switch cmdTx.Tx.State {
					case types.Tx_state_commited:
					case types.Tx_state_mined: //入块
						{
							var blockin TransactionBlockin
							blockin.AssetID = assetProperty.ID
							blockin.AssetName = assetProperty.Name
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
							status.AssetID = assetProperty.ID
							status.AssetName = assetProperty.Name
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
							status.AssetID = assetProperty.ID
							status.AssetName = assetProperty.Name
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

	_, err := db.Exec("insert transaction_blockin (asset_id, hash, status, miner_fee, blockin_height, blockin_time, order_id)"+
		" values (?, ?, ?, ?, ?, ?, ?);",
		blockin.AssetID, blockin.Hash, blockin.Status, blockin.MinerFee, blockin.BlockinHeight,
		time.Unix(blockin.BlockinTime, 0).UTC().Format(TimeFormat),
		blockin.OrderID)

	if err != nil {
		return err
	}

	return a.preSettlement(blockin, transfer)
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
		userAddress, ok := mysqlpool.QueryAllUserAddress()[blockin.AssetName+"_"+detail.Address]
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

	db.Exec("update transaction_blockin set status = ? where asset_id = ? and hash = ?;",
		status.Status, status.AssetID, status.Hash)

	rows, _ := db.Query("select asset_id, address, trans_type, amount, hash, detail_id from transaction_detail"+
		" where asset_id = ? and hash = ?;",
		status.AssetID, status.Hash)

	var detail TransactionDetail
	for rows.Next() {
		err := rows.Scan(&detail.AssetID, &detail.Address, &detail.TransType, &detail.Amount,
			&detail.Hash, &detail.DetailID)
		if err == nil {
			userAddress, ok := mysqlpool.QueryAllUserAddress()[blockin.AssetName+"_"+detail.Address]

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
