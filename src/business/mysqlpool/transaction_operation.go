package mysqlpool

import (
	. "business/def"
	"errors"
	"time"
)

func AddUserOrder(userKey string, userOrderID string, orderID string) error {
	db := Get()
	_, err := db.Exec("insert user_order (user_key, user_order_id, order_id) values (?, ?, ?)",
		userKey, userOrderID, orderID)
	return err
}

func RemoveUserOrder(userKey string, userOrderID string) error {
	db := Get()
	_, err := db.Exec("delete from user_order where user_key = ? and user_order_id = ?",
		userKey, userOrderID)
	return err
}

func WithDrawalOrder(userKey string, assetName string, address string, amount float64, payFee float64,
	uuID string, userOrderID string) error {
	nowTM := time.Now().UTC().Format(TimeFormat)

	tx, err := Get().Begin()
	if err != nil {
		return err
	}

	ret, err := tx.Exec("update user_account set available_amount = available_amount - ?, frozen_amount = frozen_amount + ?,"+
		" update_time = ? where user_key = ? and asset_name = ? and available_amount >= ?;",
		amount+payFee, amount+payFee, nowTM, userKey, assetName, amount+payFee)

	if err != nil {
		tx.Rollback()
		return err
	}

	rows, err := ret.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rows < 1 {
		tx.Rollback()
		return errors.New("update user_account rows < 1")
	}

	ret, err = tx.Exec("insert withdrawal_order (user_key, order_id, user_order_id, asset_name, address, amount, pay_fee, create_time) "+
		"values (?, ?, ?, ?, ?, ?, ?, ?)",
		userKey, uuID, userOrderID, assetName, address, amount, payFee, nowTM)

	if err != nil {
		tx.Rollback()
		return err
	}

	rows, err = ret.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rows < 1 {
		tx.Rollback()
		return errors.New("insert withdrawal_order row < 1")
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
