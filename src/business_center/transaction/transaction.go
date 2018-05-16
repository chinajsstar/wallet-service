package transaction

import (
	"blockchain_server/types"
	. "business_center/def"
	"business_center/mysqlpool"
	"encoding/json"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"github.com/satori/go.uuid"
	"time"
)

func Blockin(blockin *TransactionBlockin, transfer *types.Transfer, callback PushMsgCallback) error {
	db := mysqlpool.Get()
	blockin.Time = int64(transfer.Time)

	//为提币订单填充Hash
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
				transNotice.Hash = blockin.Hash
				db.Exec("update withdrawal_order set hash = ? where order_id = ?;", blockin.Hash, blockin.OrderID)
				SendTransactionNotic(&transNotice, callback)
			}
		}
	}

	_, err := db.Exec("insert transaction_blockin (asset_name, hash, status, miner_fee, blockin_height,"+
		" blockin_time, confirm_height, confirm_time, order_id) values (?, ?, ?, ?, ?, ?, ?, ?, ?);",
		blockin.AssetName, blockin.Hash, blockin.Status, blockin.MinerFee, blockin.BlockinHeight,
		time.Unix(blockin.Time, 0).UTC().Format(TimeFormat), blockin.BlockinHeight,
		time.Unix(blockin.Time, 0).UTC().Format(TimeFormat), blockin.OrderID)
	if err != nil {
		return nil
	}

	//地址帐单冻结处理
	preSettlement(blockin, transfer, callback)
	return err
}

func Confirm(blockin *TransactionBlockin, transfer *types.Transfer, callback PushMsgCallback) error {
	db := mysqlpool.Get()
	blockin.Time = time.Now().Unix()

	//结算提币订单
	if len(blockin.OrderID) > 0 {
		if ret, err := db.Exec("update withdrawal_order set status = 1"+
			" where order_id = ? and status = 0", blockin.OrderID); err == nil {
			if affectedRows, _ := ret.RowsAffected(); affectedRows > 0 {
				row := db.QueryRow("select user_key, asset_name, address, amount, pay_fee, hash"+
					" from withdrawal_order where order_id = ?", blockin.OrderID)
				transNotice := TransactionNotice{
					MsgID:         0,
					Type:          TypeWithdrawal,
					Status:        StatusConfirm,
					BlockinHeight: blockin.BlockinHeight,
					OrderID:       blockin.OrderID,
					Time:          blockin.Time,
				}
				err := row.Scan(&transNotice.UserKey, &transNotice.AssetName, &transNotice.Address, &transNotice.Amount,
					&transNotice.PayFee, &transNotice.Hash)
				if err == nil {
					_, err := db.Exec("update user_account set frozen_amount = frozen_amount - ?, update_time = ? where user_key = ? and asset_name = ?;",
						transNotice.Amount+transNotice.PayFee, time.Unix(blockin.Time, 0).UTC().Format(TimeFormat), transNotice.UserKey, transNotice.AssetName)
					if err != nil {
						l4g.Error(err.Error())
					}
					SendTransactionNotic(&transNotice, callback)
				}
			}
		}
	}

	_, err := db.Exec("insert transaction_status (asset_name, hash, status, confirm_height, confirm_time,"+
		" update_time, order_id) values (?, ?, ?, ?, ?, ?, ?)",
		blockin.AssetName, blockin.Hash, blockin.Status, transfer.ConfirmatedHeight,
		time.Unix(blockin.Time, 0).UTC().Format(TimeFormat),
		time.Unix(blockin.Time, 0).UTC().Format(TimeFormat), blockin.OrderID)
	if err != nil {
		return nil
	}

	db.Exec("update transaction_blockin set status = ?, confirm_height = ?, confirm_time = ?"+
		" where asset_name = ? and hash = ?",
		blockin.Status, transfer.ConfirmatedHeight, time.Unix(blockin.Time, 0).UTC().Format(TimeFormat),
		blockin.AssetName, blockin.Hash)

	//地址帐单结算
	finSettlement(blockin, transfer, callback)
	return nil
}

func preSettlement(blockin *TransactionBlockin, transfer *types.Transfer, callback PushMsgCallback) {
	db := mysqlpool.Get()
	fn := func(assetName string, address string, transType string, amount int64, hash string) {
		uuID := GenerateUUID("")
		if userAddress, ok := mysqlpool.QueryUserAddressByNameAddress(blockin.AssetName, address); ok {
			db.Exec("update user_address set available_amount = available_amount + ?, update_time = ?"+
				" where asset_name = ? and address = ?;",
				amount, time.Unix(blockin.Time, 0).UTC().Format(TimeFormat), assetName, address)

			if userAddress.UserClass == 0 && transType == "to" {
				transNotice := TransactionNotice{
					MsgID:         0,
					UserKey:       userAddress.UserKey,
					Hash:          hash,
					AssetName:     assetName,
					Address:       address,
					Amount:        amount,
					Type:          TypeDeposit,
					Status:        StatusBlockin,
					BlockinHeight: blockin.BlockinHeight,
					OrderID:       "DP" + uuID,
					Time:          int64(transfer.Time),
				}
				SendTransactionNotic(&transNotice, callback)
			}
		}
		db.Exec("insert transaction_detail (asset_name, address, trans_type, amount, hash, detail_id) "+
			" values (?, ?, ?, ?, ?, ?)",
			assetName, address, transType, amount, hash, uuID)
	}

	if transfer.IsTokenTx() {
		fn(transfer.Token.Symbol, transfer.From, "from", -int64(transfer.Value), transfer.Tx_hash)
		fn(transfer.Token.Symbol, transfer.To, "to", int64(transfer.Value), transfer.Tx_hash)
		fn(blockin.AssetName, transfer.Token.Address, "miner_fee", -int64(transfer.Value), transfer.Tx_hash)
	} else {
		fn(blockin.AssetName, transfer.From, "from", -int64(transfer.Value), transfer.Tx_hash)
		fn(blockin.AssetName, transfer.To, "to", int64(transfer.Value), transfer.Tx_hash)
		fn(blockin.AssetName, transfer.From, "miner_fee", -int64(transfer.Fee), transfer.Tx_hash)
	}
}

func finSettlement(blockin *TransactionBlockin, transfer *types.Transfer, callback PushMsgCallback) {
	db := mysqlpool.Get()
	fn := func(assetName string, address string, transType string, amount int64, hash string) {
		if userAddress, ok := mysqlpool.QueryUserAddressByNameAddress(blockin.AssetName, address); ok {
			if userAddress.UserClass == 0 && transType == "to" {
				//充值帐户余额修改
				db.Exec("update user_account set available_amount = available_amount + ?,"+
					" update_time = ? where user_key = ? and asset_name = ?",
					amount, time.Unix(blockin.Time, 0).UTC().Format(TimeFormat), userAddress.UserKey, assetName)

				transNotice := TransactionNotice{
					MsgID:         0,
					UserKey:       userAddress.UserKey,
					Hash:          hash,
					AssetName:     assetName,
					Address:       address,
					Amount:        amount,
					Type:          TypeDeposit,
					Status:        StatusConfirm,
					BlockinHeight: blockin.BlockinHeight,
					OrderID:       blockin.OrderID,
					Time:          blockin.Time,
				}
				SendTransactionNotic(&transNotice, callback)
			}
		}
	}

	if transfer.IsTokenTx() {
		fn(transfer.Token.Symbol, transfer.From, "from", -int64(transfer.Value), transfer.Tx_hash)
		fn(transfer.Token.Symbol, transfer.To, "to", int64(transfer.Value), transfer.Tx_hash)
		fn(blockin.AssetName, transfer.Token.Address, "miner_fee", -int64(transfer.Value), transfer.Tx_hash)
	} else {
		fn(blockin.AssetName, transfer.From, "from", -int64(transfer.Value), transfer.Tx_hash)
		fn(blockin.AssetName, transfer.To, "to", int64(transfer.Value), transfer.Tx_hash)
		fn(blockin.AssetName, transfer.From, "miner_fee", -int64(transfer.Fee), transfer.Tx_hash)
	}
}

func SendTransactionNotic(transNotice *TransactionNotice, callback PushMsgCallback) error {
	db := mysqlpool.Get()
	if len(transNotice.OrderID) <= 0 {
		row := db.QueryRow("select order_id from transaction_notice"+
			" where user_key = ? and asset_name = ? and hash = ? and order_id <> ''",
			transNotice.UserKey, transNotice.AssetName, transNotice.Hash)
		row.Scan(&transNotice.OrderID)
	}

	ret, err := db.Exec("insert into transaction_notice (user_key, msg_id,"+
		" trans_type, status, blockin_height, asset_name, address, amount, pay_fee, hash, order_id, time)"+
		" select ?, count(*)+1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? from transaction_notice where user_key = ?",
		transNotice.UserKey, transNotice.Type, transNotice.Status, transNotice.BlockinHeight, transNotice.AssetName,
		transNotice.Address, transNotice.Amount, transNotice.PayFee, transNotice.Hash, transNotice.OrderID,
		time.Unix(transNotice.Time, 0).Format(TimeFormat), transNotice.UserKey)
	if err != nil {
		return err
	}

	insertID, err := ret.LastInsertId()
	if err != nil {
		return err
	}

	row := db.QueryRow("select msg_id from transaction_notice where id = ?", insertID)
	row.Scan(&transNotice.MsgID)

	data, err := json.Marshal(transNotice)
	if err != nil {
		return err
	}

	if callback != nil {
		callback(transNotice.UserKey, string(data))
	}

	return nil
}

func GenerateUUID(prefix string) string {
	uID, _ := uuid.NewV4()
	return fmt.Sprintf("%s%s%X", prefix, time.Now().UTC().Format("20060102150405"), uID.Bytes())
}
