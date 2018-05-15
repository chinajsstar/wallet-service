package mysqlpool

import (
	. "business_center/def"
	"errors"
	"time"
)

func WithDrawalOrder(userKey string, assetName string, address string, amount int64, payFee int64, uuID string) error {
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

	ret, err = tx.Exec("insert withdrawal_order (order_id, user_key, asset_name, address, amount, pay_fee, create_time) "+
		"values (?, ?, ?, ?, ?, ?, ?);",
		uuID, userKey, assetName, address, amount, payFee, nowTM)

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
