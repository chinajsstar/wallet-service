package transaction

import (
	"bastionpay_api/api/v1"
	"blockchain_server/types"
	. "business_center/def"
	"business_center/mysqlpool"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"strconv"
	"time"
)

func Blockin(blockin *TransactionBlockin, transfer *types.Transfer, callback PushMsgCallback) error {
	db := mysqlpool.Get()
	blockin.Time = int64(transfer.Time)

	//为提币订单填充Hash
	if len(blockin.OrderID) > 0 {
		row := db.QueryRow("select user_key, asset_name, address, amount, pay_fee, hash"+
			" from withdrawal_order where order_id = ?", blockin.OrderID)
		transNotice := TransactionNotice{
			MsgID:         0,
			TransType:     TypeWithdrawal,
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
				transNotice.MinerFee = blockin.MinerFee
				db.Exec("update withdrawal_order set hash = ?, miner_fee = ? where order_id = ?",
					blockin.Hash, blockin.MinerFee, blockin.OrderID)
				row := db.QueryRow("select available_amount from user_account where user_key = ? and asset_name = ?",
					transNotice.UserKey, transNotice.AssetName)
				row.Scan(&transNotice.Balance)
				SendTransactionNotic(&transNotice, callback)
				AddProfitBill(&transNotice)
				AddTransactionBill(&transNotice)
				AddTransactionBillDaily(&transNotice)
			}
		}
	}

	_, err := db.Exec("insert transaction_blockin (asset_name, hash, status, miner_fee, blockin_height,"+
		" blockin_time, confirm_height, confirm_time, order_id) values (?, ?, ?, ?, ?, ?, ?, ?, ?)",
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
				row := db.QueryRow("select user_key, asset_name, address, amount, pay_fee, miner_fee, hash, user_order_id"+
					" from withdrawal_order where order_id = ?", blockin.OrderID)
				transNotice := TransactionNotice{
					MsgID:         0,
					TransType:     TypeWithdrawal,
					Status:        StatusConfirm,
					BlockinHeight: blockin.BlockinHeight,
					OrderID:       blockin.OrderID,
					Time:          blockin.Time,
				}
				var userOrderID string
				err := row.Scan(&transNotice.UserKey, &transNotice.AssetName, &transNotice.Address, &transNotice.Amount,
					&transNotice.PayFee, &transNotice.MinerFee, &transNotice.Hash, &userOrderID)
				if err == nil {
					tx, _ := db.Begin()
					tx.Exec("update user_account set frozen_amount = frozen_amount - ?, update_time = ?"+
						" where user_key = ? and asset_name = ?;",
						transNotice.Amount+transNotice.PayFee, time.Unix(blockin.Time, 0).UTC().Format(TimeFormat),
						transNotice.UserKey, transNotice.AssetName)
					row := tx.QueryRow("select available_amount from user_account where user_key = ? and asset_name = ?",
						transNotice.UserKey, transNotice.AssetName)
					row.Scan(&transNotice.Balance)
					tx.Commit()
					mysqlpool.RemoveUserOrder(transNotice.UserKey, userOrderID)
					SendTransactionNotic(&transNotice, callback)
					AddProfitBill(&transNotice)
					AddTransactionBill(&transNotice)
					AddTransactionBillDaily(&transNotice)
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
	fn := func(assetName string, address string, transType string, amount float64, hash string) {
		uuID := GenerateUUID("")
		if userAddress, ok := mysqlpool.QueryUserAddressByNameAddress(assetName, address); ok {
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
					TransType:     TypeDeposit,
					Status:        StatusBlockin,
					BlockinHeight: blockin.BlockinHeight,
					OrderID:       "DP" + uuID,
					Time:          int64(transfer.Time),
				}
				row := db.QueryRow("select available_amount from user_account where user_key = ? and asset_name = ?",
					transNotice.UserKey, transNotice.AssetName)
				row.Scan(&transNotice.Balance)
				SendTransactionNotic(&transNotice, callback)
				AddProfitBill(&transNotice)
				AddTransactionBill(&transNotice)
				AddTransactionBillDaily(&transNotice)
			}
		}
		db.Exec("insert transaction_detail (asset_name, address, trans_type, amount, hash, detail_id) "+
			" values (?, ?, ?, ?, ?, ?)",
			assetName, address, transType, amount, hash, uuID)
	}

	if transfer.IsTokenTx() {
		fn(transfer.TokenTx.Symbol(), transfer.TokenTx.From, "from", -transfer.TokenTx.Value, transfer.Tx_hash)
		fn(transfer.TokenTx.Symbol(), transfer.TokenTx.To, "to", transfer.TokenTx.Value, transfer.Tx_hash)
		fn(blockin.AssetName, transfer.From, "miner_fee", -transfer.Fee, transfer.Tx_hash)
	} else {
		fn(blockin.AssetName, transfer.From, "from", -transfer.Value, transfer.Tx_hash)
		fn(blockin.AssetName, transfer.To, "to", transfer.Value, transfer.Tx_hash)
		fn(blockin.AssetName, transfer.From, "miner_fee", -transfer.Fee, transfer.Tx_hash)
	}
}

func finSettlement(blockin *TransactionBlockin, transfer *types.Transfer, callback PushMsgCallback) {
	db := mysqlpool.Get()
	fn := func(assetName string, address string, transType string, amount float64, hash string) {
		if userAddress, ok := mysqlpool.QueryUserAddressByNameAddress(assetName, address); ok {
			if userAddress.UserClass == 0 && transType == "to" {
				//充值帐户余额修改
				tx, _ := db.Begin()
				tx.Exec("update user_account set available_amount = available_amount + ?,"+
					" update_time = ? where user_key = ? and asset_name = ?",
					amount, time.Unix(blockin.Time, 0).UTC().Format(TimeFormat), userAddress.UserKey, assetName)
				transNotice := TransactionNotice{
					MsgID:         0,
					UserKey:       userAddress.UserKey,
					Hash:          hash,
					AssetName:     assetName,
					Address:       address,
					Amount:        amount,
					TransType:     TypeDeposit,
					Status:        StatusConfirm,
					BlockinHeight: blockin.BlockinHeight,
					OrderID:       "",
					Time:          blockin.Time,
				}
				row := tx.QueryRow("select available_amount from user_account where user_key = ? and asset_name = ?",
					transNotice.UserKey, transNotice.AssetName)
				row.Scan(&transNotice.Balance)
				tx.Commit()
				SendTransactionNotic(&transNotice, callback)
				AddProfitBill(&transNotice)
				AddTransactionBill(&transNotice)
				AddTransactionBillDaily(&transNotice)
			}
		}
	}

	if transfer.IsTokenTx() {
		fn(transfer.TokenTx.Symbol(), transfer.TokenTx.From, "from", -transfer.TokenTx.Value, transfer.Tx_hash)
		fn(transfer.TokenTx.Symbol(), transfer.TokenTx.To, "to", transfer.TokenTx.Value, transfer.Tx_hash)
		fn(blockin.AssetName, transfer.From, "miner_fee", -transfer.Fee, transfer.Tx_hash)
	} else {
		fn(blockin.AssetName, transfer.From, "from", -transfer.Value, transfer.Tx_hash)
		fn(blockin.AssetName, transfer.To, "to", transfer.Value, transfer.Tx_hash)
		fn(blockin.AssetName, transfer.From, "miner_fee", -transfer.Fee, transfer.Tx_hash)
	}
}

func AddTransactionBill(transNotice *TransactionNotice) error {
	tx, err := mysqlpool.Get().Begin()
	if err != nil {
		return err
	}

	ret, err := tx.Exec("update transaction_bill set status = ?, balance = ?, time = ? where status = 0 and user_key = ?"+
		" and asset_name = ? and order_id = ?",
		transNotice.Status, transNotice.Balance, time.Unix(transNotice.Time, 0).UTC().Format(TimeFormat),
		transNotice.UserKey, transNotice.AssetName, transNotice.OrderID)
	if err != nil {
		return err
	}

	rowAffected, err := ret.RowsAffected()
	if err != nil {
		return err
	}

	if rowAffected > 0 {
		return tx.Commit()
	}

	_, err = tx.Exec("insert transaction_bill (user_key, trans_type, status, blockin_height, asset_name, address, amount, "+
		"pay_fee, miner_fee, balance, hash, order_id, time) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		transNotice.UserKey, transNotice.TransType, transNotice.Status, transNotice.BlockinHeight, transNotice.AssetName,
		transNotice.Address, transNotice.Amount, transNotice.PayFee, transNotice.MinerFee, transNotice.Balance, transNotice.Hash,
		transNotice.OrderID, time.Unix(transNotice.Time, 0).UTC().Format(TimeFormat))
	if err != nil {
		return err
	}

	return tx.Commit()
}

func AddTransactionBillDaily(transNotice *TransactionNotice) error {
	if transNotice.Status == 1 {
		period, err := strconv.Atoi(time.Unix(transNotice.Time, 0).Format(DateFormat))
		if err != nil {
			return err
		}

		db := mysqlpool.Get()
		timeF := time.Unix(transNotice.Time, 0).UTC().Format(TimeFormat)
		if transNotice.TransType == 0 {
			ret, err := db.Exec("update transaction_bill_daily set sum_dp_amount = sum_dp_amount + ?,"+
				" pre_balance = if(pre_time > ?, ?, pre_balance),"+
				" last_balance = if(last_time < ?, ?, last_balance),"+
				" pre_time = if(pre_time > ?, ?, pre_time),"+
				" last_time = if(last_time < ?, ?, last_time)"+
				" where period = ? and user_key = ? and asset_name = ?",
				transNotice.Amount, timeF, transNotice.Balance, timeF, transNotice.Balance, timeF, timeF, timeF, timeF,
				period, transNotice.UserKey, transNotice.AssetName)
			if err != nil {
				return err
			}

			rowAffected, err := ret.RowsAffected()
			if err != nil {
				return err
			}

			if rowAffected > 0 {
				return nil
			}

			_, err = db.Exec("insert transaction_bill_daily (period, user_key, asset_name, sum_dp_amount,"+
				" sum_wd_amount, sum_pay_fee, sum_miner_fee, pre_balance, last_balance, pre_time, last_time)"+
				" values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				period, transNotice.UserKey, transNotice.AssetName, transNotice.Amount, 0, 0, 0,
				transNotice.Balance-transNotice.Amount, transNotice.Balance, timeF, timeF)
			if err != nil {
				return err
			}
		} else {
			ret, err := db.Exec("update transaction_bill_daily set sum_wd_amount = sum_wd_amount + ?,"+
				" sum_pay_fee = sum_pay_fee + ?,"+
				" sum_miner_fee = sum_miner_fee + ?,"+
				" pre_balance = if(pre_time > ?, ?, pre_balance),"+
				" last_balance = if(last_time < ?, ?, last_balance),"+
				" pre_time = if(pre_time > ?, ?, pre_time),"+
				" last_time = if(last_time < ?, ?, last_time)"+
				" where period = ? and user_key = ? and asset_name = ?",
				transNotice.Amount, transNotice.PayFee, transNotice.MinerFee, timeF, transNotice.Balance, timeF,
				transNotice.Balance, timeF, timeF, timeF, timeF, period, transNotice.UserKey, transNotice.AssetName)
			if err != nil {
				return err
			}

			rowAffected, err := ret.RowsAffected()
			if err != nil {
				return err
			}

			if rowAffected > 0 {
				return nil
			}

			_, err = db.Exec("insert transaction_bill_daily (period, user_key, asset_name, sum_dp_amount,"+
				" sum_wd_amount, sum_pay_fee, sum_miner_fee, pre_balance, last_balance, pre_time, last_time)"+
				" values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				period, transNotice.UserKey, transNotice.AssetName, 0, transNotice.Amount, transNotice.PayFee,
				transNotice.MinerFee, transNotice.Balance-transNotice.Amount, transNotice.Balance, timeF, timeF)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func SendTransactionNotic(transNotice *TransactionNotice, callback PushMsgCallback) error {
	db := mysqlpool.Get()
	if len(transNotice.OrderID) <= 0 {
		if transNotice.TransType == TypeDeposit {
			row := db.QueryRow("select order_id from transaction_notice"+
				" where user_key = ? and asset_name = ? and hash = ? and order_id like 'DP%'",
				transNotice.UserKey, transNotice.AssetName, transNotice.Hash)
			row.Scan(&transNotice.OrderID)
		} else if transNotice.TransType == TypeWithdrawal {
			row := db.QueryRow("select order_id from transaction_notice"+
				" where user_key = ? and asset_name = ? and hash = ? and order_id like 'WD%'",
				transNotice.UserKey, transNotice.AssetName, transNotice.Hash)
			row.Scan(&transNotice.OrderID)
		}
	}

	ret, err := db.Exec("insert into transaction_notice (user_key, msg_id, trans_type, status, blockin_height,"+
		" asset_name, address, amount, pay_fee, miner_fee, balance, hash, order_id, time)"+
		" select ?, count(*)+1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? from transaction_notice where user_key = ?",
		transNotice.UserKey, transNotice.TransType, transNotice.Status, transNotice.BlockinHeight, transNotice.AssetName,
		transNotice.Address, transNotice.Amount, transNotice.PayFee, transNotice.MinerFee, transNotice.Balance,
		transNotice.Hash, transNotice.OrderID, time.Unix(transNotice.Time, 0).UTC().Format(TimeFormat), transNotice.UserKey)
	if err != nil {
		return err
	}

	insertID, err := ret.LastInsertId()
	if err != nil {
		return err
	}

	row := db.QueryRow("select msg_id from transaction_notice where id = ?", insertID)
	row.Scan(&transNotice.MsgID)

	data := v1.PushTransactionMessage{
		MsgID:         transNotice.MsgID,
		TransType:     transNotice.TransType,
		Status:        transNotice.Status,
		BlockinHeight: transNotice.BlockinHeight,
		AssetName:     transNotice.AssetName,
		Address:       transNotice.Address,
		Amount:        transNotice.Amount,
		PayFee:        transNotice.PayFee,
		Balance:       transNotice.Balance,
		Hash:          transNotice.Hash,
		OrderID:       transNotice.OrderID,
		Time:          transNotice.Time,
	}

	pack, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if callback != nil {
		callback(transNotice.UserKey, string(pack))
	}

	return nil
}

func AddProfitBill(transNotice *TransactionNotice) error {
	if transNotice.Status == 1 {
		tx, err := mysqlpool.Get().Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec("insert profit_bill (profit_user_key, user_key, trans_type, asset_name, order_id,"+
			" hash, amount, pay_fee, miner_fee, profit, time) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			"", transNotice.UserKey, transNotice.TransType, transNotice.AssetName, transNotice.OrderID,
			transNotice.Hash, transNotice.Amount, transNotice.PayFee, transNotice.MinerFee,
			transNotice.PayFee-transNotice.MinerFee, time.Unix(transNotice.Time, 0).UTC().Format(TimeFormat))
		if err != nil {
			return err
		}

		ret, err := tx.Exec("update profit_bill_daily set sum_profit = sum_profit + ?, time = ?"+
			" where period = ? and profit_user_key = ? and asset_name = ?",
			transNotice.PayFee-transNotice.MinerFee, time.Unix(transNotice.Time, 0).UTC().Format(TimeFormat),
			time.Unix(transNotice.Time, 0).UTC().Format(DateFormat), "", transNotice.AssetName)
		if err != nil {
			return err
		}

		rowAffected, err := ret.RowsAffected()
		if err != nil {
			return err
		}

		if rowAffected > 0 {
			return tx.Commit()
		}

		_, err = tx.Exec("insert profit_bill_daily(period, profit_user_key, asset_name, sum_profit, time)"+
			" values (?, ?, ?, ?, ?)",
			time.Unix(transNotice.Time, 0).UTC().Format(DateFormat), "", transNotice.AssetName,
			transNotice.PayFee-transNotice.MinerFee, time.Unix(transNotice.Time, 0).UTC().Format(TimeFormat))
		if err != nil {
			return err
		}

		return tx.Commit()
	}
	return nil
}

func GenerateUUID(prefix string) string {
	uID, _ := uuid.NewV4()
	return fmt.Sprintf("%s%s%X", prefix, time.Now().UTC().Format("20060102150405"), uID.Bytes())
}
