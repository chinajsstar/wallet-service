package mysqlpool

import (
	. "business_center/def"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"time"
)

func QueryUserAddress(queryMap map[string]interface{}) ([]UserAddress, bool) {
	sqls := "select user_key,user_class,asset_name,address,private_key,available_amount,frozen_amount,enabled," +
		" unix_timestamp(create_time), unix_timestamp(allocation_time), unix_timestamp(update_time) from user_address" +
		" where true"

	userAddress := make([]UserAddress, 0)
	params := make([]interface{}, 0)

	if len(queryMap) > 0 {
		sqls += andConditions(queryMap, &params)
		sqls += andPagination(queryMap, &params)
	}

	db := Get()
	rows, err := db.Query(sqls, params...)
	if err != nil {
		fmt.Println(err.Error())
		return userAddress, len(userAddress) > 0
	}

	var data UserAddress
	for rows.Next() {
		err := rows.Scan(&data.UserKey, &data.UserClass, &data.AssetName, &data.Address, &data.PrivateKey,
			&data.AvailableAmount, &data.FrozenAmount, &data.Enabled, &data.CreateTime, &data.AllocationTime,
			&data.UpdateTime)
		if err == nil {
			userAddress = append(userAddress, data)
		}
	}
	return userAddress, len(userAddress) > 0
}

func QueryUserAddressCount(queryMap map[string]interface{}) int {
	sqls := "select count(*) from user_address" +
		" where true"

	count := 0
	params := make([]interface{}, 0)

	if len(queryMap) > 0 {
		sqls += andConditions(queryMap, &params)
	}

	db := Get()
	db.QueryRow(sqls, params...).Scan(&count)
	return count
}

func QueryUserAddressByNameAddress(assetName string, address string) (UserAddress, bool) {
	queryMap := make(map[string]interface{})
	queryMap["asset_name"] = assetName
	queryMap["address"] = address
	if userAddress, ok := QueryUserAddress(queryMap); ok {
		return userAddress[0], true
	}
	return UserAddress{}, false
}

func QueryPayAddress(assetName string) (UserAddress, bool) {
	db := Get()
	row := db.QueryRow("select user_key,user_class,asset_name,address,private_key,available_amount,frozen_amount,"+
		"enabled, unix_timestamp(create_time), unix_timestamp(update_time) from pay_address_view where asset_name = ?;", assetName)
	var userAddress UserAddress
	err := row.Scan(&userAddress.UserKey, &userAddress.UserClass, &userAddress.AssetName, &userAddress.Address,
		&userAddress.PrivateKey, &userAddress.AvailableAmount, &userAddress.FrozenAmount, &userAddress.Enabled,
		&userAddress.CreateTime, &userAddress.UpdateTime)
	if err != nil {
		return userAddress, false
	}
	return userAddress, true
}

func SetPayAddress(assetName string, address string) error {
	db := Get()
	row := db.QueryRow("select private_key from user_address"+
		" where user_class = 1 and asset_name = ? and address = ?;", assetName, address)

	var privateKey string
	err := row.Scan(&privateKey)
	if err != nil {
		return err
	}

	_, err = db.Exec("insert pay_address (asset_name, address, private_key) values (?, ?, ?);",
		assetName, address, privateKey)
	if err != nil {
		_, err := db.Exec("update pay_address set address = ?, private_key = ? where asset_name = ?;",
			address, privateKey, assetName)
		if err != nil {
			return err
		}
	}

	return nil
}

func AddUserAddress(userAddress []UserAddress) error {
	tx, err := Get().Begin()
	if err != nil {
		_, errMsg := CheckError(ErrorDataBase, err.Error())
		l4g.Error(errMsg)
		return err
	}

	for _, v := range userAddress {
		_, err := tx.Exec("insert user_address (user_key, user_class, asset_name, address, private_key,"+
			" available_amount, frozen_amount, enabled, create_time, allocation_time, update_time)"+
			" values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);",
			v.UserKey, v.UserClass, v.AssetName, v.Address, v.PrivateKey, v.AvailableAmount, v.FrozenAmount, v.Enabled,
			time.Unix(v.CreateTime, 0).UTC().Format(TimeFormat),
			time.Unix(v.UpdateTime, 0).UTC().Format(TimeFormat),
			time.Unix(v.UpdateTime, 0).UTC().Format(TimeFormat))
		if err != nil {
			tx.Rollback()
			_, errMsg := CheckError(ErrorDataBase, err.Error())
			l4g.Error(errMsg)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		_, errMsg := CheckError(ErrorDataBase, err.Error())
		l4g.Error(errMsg)
		return err
	}

	return nil
}