package address

import (
	"blockchain_server/service"
	"blockchain_server/types"
	"business_center/basicdata"
	. "business_center/def"
	"business_center/mysqlpool"
	"business_center/redispool"
	"context"
	"fmt"
	"strings"
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
		userAddress.UserID = userID
		userAddress.UserClass = userClass
		userAddress.AssetID = assetID
		userAddress.AssetName = assetName
		userAddress.Address = strings.ToLower(accounts[0].Address)
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
		_, err := tx.Exec("insert user_address (user_id, asset_id, address, private_key, available_amount, frozen_amount, "+
			"enabled, create_time, update_time) values (?, ?, ?, ?, ?, ?, ?, ?, ?);",
			v.UserID, v.AssetID, v.Address, v.PrivateKey, v.AvailableAmount, v.FrozenAmount, v.Enabled,
			time.Unix(v.CreateTime, 0).UTC().Format("2006-01-02 15:04:05"),
			time.Unix(v.UpdateTime, 0).UTC().Format("2006-01-02 15:04:05"))
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		addresses = append(addresses, v.Address)
	}
	tx.Commit()
	basicdata.Get().AddUserAddressMap(userAddress)

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
					assetProperty, ok := basicdata.Get().GetAllAssetPropertyMap()[rct.Coin_name]
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
					assetProperty, ok := basicdata.Get().GetAllAssetPropertyMap()[cmdTx.Coinname]
					if !ok {
						continue
					}

					switch cmdTx.Tx.State {
					case types.Tx_state_mined: //入块
						{
							var blockin TransactionBlockin
							blockin.AssetID = assetProperty.ID
							blockin.AssetName = assetProperty.Name
							blockin.Hash = cmdTx.Tx.Tx_hash
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
	_, err := db.Exec("insert transaction_blockin "+
		"(asset_id, hash, blockin_height, blockin_time, order_id) values (?, ?, ?, ?, ?);",
		blockin.AssetID, blockin.Hash, blockin.BlockinHeight,
		time.Unix(blockin.BlockinTime, 0).UTC().Format("2006-01-02 15:04:05"),
		blockin.OrderID)

	if err != nil {
		return err
	}

	return a.preSettlement(blockin, transfer)
}

func (a *Address) preSettlement(blockin *TransactionBlockin, transfer *types.Transfer) error {

	switch blockin.AssetName {
	case "btc":
		{

		}
	case "eth":
		{
			var detail TransactionDetail
			blockin.Detail = make([]TransactionDetail, 0)

			//from
			detail.Address = transfer.From
			detail.Amount = -int64(transfer.Value)
			detail.TransType = "from"
			blockin.Detail = append(blockin.Detail, detail)

			//to
			detail.Address = transfer.To
			detail.Amount = int64(transfer.Value)
			detail.TransType = "to"
			blockin.Detail = append(blockin.Detail, detail)

			//gas
			detail.Address = transfer.From
			detail.Amount = -int64(transfer.Gase)
			detail.TransType = "gas"
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

		userAddress, ok := basicdata.Get().GetAllUserAddressMap()[blockin.AssetName+"_"+detail.Address]
		if ok {
			_, err := Tx.Exec("update user_address set available_amount = available_amount + ?, update_time = now() "+
				" where asset_id = ? and address = ?;",
				detail.Amount, userAddress.AssetID, userAddress.Address)
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		_, err := Tx.Exec("insert transaction_detail "+
			"(asset_id, hash, address, trans_type, amount) "+
			"values (?, ?, ?, ?, ?);",
			blockin.AssetID, blockin.Hash, detail.Address, detail.TransType, detail.Amount)

		if err != nil {
			continue
		}
	}

	Tx.Commit()

	return nil
}

func (a *Address) transactionFinish(status *TransactionStatus, transfer *types.Transfer) error {
	db := mysqlpool.Get()
	_, err := db.Exec("insert transaction_status "+
		"(asset_id, hash, status, confirm_height, confirm_time, update_time, order_id) "+
		"values (?, ?, ?, ?, ?, ?, ?);",
		status.AssetID, status.Hash, status.Status, status.ConfirmHeight,
		time.Unix(status.ConfirmTime, 0).UTC().Format("2006-01-02 15:04:05"),
		time.Unix(status.UpdateTime, 0).UTC().Format("2006-01-02 15:04:05"),
		status.OrderID)

	if err != nil {
		return err
	}

	if status.Status == 1 {
		var blockin TransactionBlockin
		blockin.AssetName = status.AssetName
		blockin.Detail = make([]TransactionDetail, 0)

		row := db.QueryRow("select asset_id, hash, blockin_height, blockin_time, order_id"+
			" from transaction_blockin where asset_id = ? and hash = ?",
			status.AssetID, status.Hash)

		if row == nil {
			return nil
		}

		row.Scan(&blockin.AssetID, &blockin.Hash, &blockin.BlockinHeight, &blockin.BlockinTime, &blockin.OrderID)
		if len(status.OrderID) > 0 {
			blockin.OrderID = status.OrderID
		}

		rows, err := db.Query("select address, trans_type, amount from transaction_detail where asset_id = ? and hash = ?;",
			status.AssetID, status.Hash)

		if err != nil {
			return err
		}

		var detail TransactionDetail
		for rows.Next() {
			rows.Scan(&detail.Address, &detail.TransType, &detail.Amount)
			blockin.Detail = append(blockin.Detail, detail)
		}

		//结算订单
		if len(blockin.OrderID) > 0 {
			row := db.QueryRow("select user_id, asset_id, amount, wallet_fee"+
				" from withdraw_order where order_id = ?", blockin.OrderID)
			if row != nil {
				var (
					userID    string
					assetID   int
					amount    int64
					walletFee int64
				)
				err = row.Scan(&userID, &assetID, &amount, &walletFee)
				if err == nil {
					db.Exec("update user_account set frozen_amount = frozen_amount - ?, update_time = now()"+
						" where user_id = ? and asset_id = ?;", amount+walletFee, userID, assetID)
				}
			}
		}

		//充值订单
		for _, v := range blockin.Detail {
			userAddress, ok := basicdata.Get().GetAllUserAddressMap()[blockin.AssetName+"_"+v.Address]
			if ok && userAddress.UserClass == 0 {
				switch blockin.AssetName {
				case "btc":
					if v.TransType == "to" || v.TransType == "gas" || v.TransType == "change" {
						db.Exec("update user_account set available_amount = available_amount + ?,"+
							" update_time = now() where user_id = ? and asset_id = ?;",
							v.Amount, userAddress.UserID, userAddress.AssetID)
					}
				case "eth":
					if v.TransType == "to" {
						db.Exec("update user_account set available_amount = available_amount + ?,"+
							" update_time = now() where user_id = ? and asset_id = ?;",
							v.Amount, userAddress.UserID, userAddress.AssetID)
					}
				}
			}
		}
	}

	return nil
}
