package mysqlpool

import (
	. "business_center/def"
	"encoding/json"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"time"
)

func QueryUserAddress(query string) ([]UserAddress, bool) {
	sqls := "select user_key,user_class,asset_name,address,private_key,available_amount,frozen_amount,enabled," +
		" unix_timestamp(create_time),unix_timestamp(update_time) from user_address" +
		" where true"

	userAddress := make([]UserAddress, 0)
	params := make([]interface{}, 0)

	if len(query) > 0 {
		var queryMap map[string]interface{}
		err := json.Unmarshal([]byte(query), &queryMap)
		if err != nil {
			return userAddress, len(userAddress) > 0
		}

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
			&data.AvailableAmount, &data.FrozenAmount, &data.Enabled, &data.CreateTime, &data.UpdateTime)
		if err == nil {
			userAddress = append(userAddress, data)
		}
	}
	return userAddress, len(userAddress) > 0
}

func QueryUserAddressCount(query string) int {
	sqls := "select count(*) from user_address" +
		" where true"

	count := 0
	params := make([]interface{}, 0)

	if len(query) > 0 {
		var queryMap map[string]interface{}
		err := json.Unmarshal([]byte(query), &queryMap)
		if err != nil {
			return count
		}
		sqls += andConditions(queryMap, &params)
	}

	db := Get()
	db.QueryRow(sqls, params...).Scan(&count)
	return count
}

func QueryUserAddressByNameAddress(assetName string, address string) (UserAddress, bool) {
	query := fmt.Sprintf("{\"asset_name\":%s, \"address\":\"%s\"}", assetName, address)
	if userAddress, ok := QueryUserAddress(query); ok {
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

func AddUserAddress(userAddress []UserAddress) error {
	tx, err := Get().Begin()
	if err != nil {
		_, errMsg := CheckError(ErrorDataBase, err.Error())
		l4g.Error(errMsg)
		return err
	}

	for _, v := range userAddress {
		_, err := tx.Exec("insert user_address (user_key, user_class, asset_name, address, private_key,"+
			" available_amount, frozen_amount, enabled, create_time, update_time) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);",
			v.UserKey, v.UserClass, v.AssetName, v.Address, v.PrivateKey, v.AvailableAmount, v.FrozenAmount, v.Enabled,
			time.Unix(v.CreateTime, 0).UTC().Format(TimeFormat),
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
