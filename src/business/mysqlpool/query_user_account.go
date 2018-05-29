package mysqlpool

import (
	. "business/def"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"time"
)

func QueryUserAccount(queryMap map[string]interface{}) ([]UserAccount, bool) {
	sqls := "select user_key,user_class,asset_name,available_amount,frozen_amount," +
		"unix_timestamp(create_time), unix_timestamp(update_time) from user_account" +
		" where true"

	userAccount := make([]UserAccount, 0)
	params := make([]interface{}, 0)

	if len(queryMap) > 0 {
		sqls += andConditions(queryMap, &params)
		sqls += andPagination(queryMap, &params)
	}

	db := Get()
	rows, err := db.Query(sqls, params...)
	defer rows.Close()
	if err != nil {
		fmt.Println(err.Error())
		return userAccount, len(userAccount) > 0
	}

	var data UserAccount
	for rows.Next() {
		err := rows.Scan(&data.UserKey, &data.UserClass, &data.AssetName, &data.AvailableAmount,
			&data.FrozenAmount, &data.CreateTime, &data.UpdateTime)
		if err == nil {
			userAccount = append(userAccount, data)
		}
	}
	return userAccount, len(userAccount) > 0
}

func QueryUserAccountCount(queryMap map[string]interface{}) int {
	sqls := "select count(*) from user_account" +
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

func QueryUserAccountRow(userKey string, assetName string) (UserAccount, bool) {
	queryMap := make(map[string]interface{})
	queryMap["user_key"] = userKey
	queryMap["asset_name"] = assetName
	if userAccount, ok := QueryUserAccount(queryMap); ok {
		return userAccount[0], ok
	}
	return UserAccount{}, false
}

func AddUserAccount(userKey string, userClass int, assetName string) error {
	db := Get()
	nowTM := time.Now().UTC().Format(TimeFormat)
	_, err := db.Exec("insert user_account (user_key, user_class, asset_name, available_amount, frozen_amount,"+
		" create_time, update_time) values (?, ?, ?, 0, 0, ?, ?);",
		userKey, userClass, assetName, nowTM, nowTM)
	if err != nil {
		count := 0
		row := db.QueryRow("select count(*) from user_account where user_key = ? and asset_name = ?;", userKey, assetName)
		err := row.Scan(&count)
		if err != nil {
			_, errMsg := CheckError(ErrorFailed, err.Error())
			l4g.Error(errMsg)
			return err
		}
		if count > 0 {
			return nil
		}
		return err
	}
	return nil
}
