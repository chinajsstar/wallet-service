package transaction

import (
	"blockchain_server/types"
	. "business_center/def"
	"business_center/mysqlpool"
	"encoding/json"
	l4g "github.com/alecthomas/log4go"
	"time"
)

func Blockin(blockin *TransactionBlockin, transfer *types.Transfer, callback PushMsgCallback) error {
	db := mysqlpool.Get()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("insert transaction_blockin (asset_name, hash, status, miner_fee, blockin_height,"+
		" blockin_time, confirm_height, confirm_time, order_id) values (?, ?, ?, ?, ?, ?, ?, ?, ?);",
		blockin.AssetName, blockin.Hash, blockin.Status, blockin.MinerFee, blockin.BlockinHeight,
		time.Unix(int64(transfer.Time), 0).UTC().Format(TimeFormat), blockin.BlockinHeight,
		time.Unix(int64(transfer.Time), 0).UTC().Format(TimeFormat), blockin.OrderID)
	if err == nil {
		//冻结帐单处理
		err = tx.Commit()
		if err != nil {
			err = tx.Rollback()
		}
	} else {
		err = tx.Rollback()
	}

	//通知提币入块消息
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
				SendTransactionNotic(&transNotice, callback)
			}
		}
	}
	return err
}

func Confirm(blockin *TransactionBlockin, transfer *types.Transfer, callback PushMsgCallback) error {
	db := mysqlpool.Get()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	nowTM := time.Now().Unix()
	_, err = tx.Exec("insert transaction_status (asset_name, hash, status, confirm_height, confirm_time,"+
		" update_time, order_id) values (?, ?, ?, ?, ?, ?, ?);",
		blockin.AssetName, blockin.Hash, blockin.Status, transfer.ConfirmatedHeight,
		time.Unix(nowTM, 0).UTC().Format(TimeFormat), time.Unix(nowTM, 0).UTC().Format(TimeFormat), blockin.OrderID)
	if err != nil {
		return tx.Rollback()
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	db.Exec("update transaction_blockin set status = ?, confirm_height = ?, confirm_time = ?"+
		" where asset_name = ? and hash = ?;",
		blockin.Status, transfer.ConfirmatedHeight, time.Unix(blockin.Time, 0).UTC().Format(TimeFormat),
		blockin.AssetName, blockin.Hash)

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
			_, err := db.Exec("update user_account set frozen_amount = frozen_amount - ?, update_time = ?"+
				" where user_key = ? and asset_name = ?;",
				transNotice.Amount+transNotice.PayFee, time.Now().UTC().Format(TimeFormat), transNotice.UserKey, transNotice.AssetName)
			if err != nil {
				l4g.Error(err.Error())
			}
			SendTransactionNotic(&transNotice, callback)
		}
	}
	return nil
}

func SendTransactionNotic(transNotice *TransactionNotice, callback PushMsgCallback) error {
	db := mysqlpool.Get()
	ret, err := db.Exec("insert into transaction_notice (user_key, msg_id,"+
		" trans_type, status, blockin_height, asset_name, address, amount, pay_fee, hash, order_id, time)"+
		" select ?, count(*)+1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? from transaction_notice where user_key = ?;",
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

	row := db.QueryRow("select msg_id from transaction_notice where id = ?;", insertID)
	row.Scan(&transNotice.MsgID)

	data, err := json.Marshal(transNotice)
	if err != nil {
		return err
	}

	callback(transNotice.UserKey, string(data))
	return nil
}
