package mysqlpool

import (
	. "business_center/def"
	"errors"
	"time"
)

func WithDrawalSet(userKey string, assetID int, address string, amount int64, walletFee int64, uuID string) error {
	nowTM := time.Now().UTC().Format(TimeFormat)

	tx, err := Get().Begin()
	if err != nil {
		return err
	}

	ret, err := tx.Exec("update user_account set available_amount = available_amount - ?, frozen_amount = frozen_amount + ?,"+
		" update_time = ? where user_key = ? and asset_id = ? and available_amount >= ?;",
		amount+walletFee, amount+walletFee, nowTM, userKey, assetID, amount+walletFee)

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

	ret, err = tx.Exec("insert withdrawal_order (order_id, user_key, asset_id, address, amount, wallet_fee, create_time) "+
		"values (?, ?, ?, ?, ?, ?, ?);",
		uuID, userKey, assetID, address, amount, walletFee, nowTM)

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
