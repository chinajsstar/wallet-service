package mysqlpool

import (
	. "business_center/def"
	"encoding/json"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"time"
)

func QueryUserAccountByJson(query string) ([]UserAccount, bool) {
	sqls := "select user_key,user_class,asset_name,available_amount,frozen_amount," +
		"unix_timestamp(create_time), unix_timestamp(update_time) from user_account" +
		" where true"

	userAccount := make([]UserAccount, 0)
	params := make([]interface{}, 0)

	if len(query) > 0 {
		var queryMap map[string]interface{}
		err := json.Unmarshal([]byte(query), &queryMap)
		if err != nil {
			return userAccount, len(userAccount) > 0
		}

		sqls += andConditions(queryMap, &params)
		sqls += andPagination(queryMap, &params)
	}

	db := Get()
	rows, err := db.Query(sqls, params...)
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

func QueryUserAccountCountByJson(query string) int {
	sqls := "select count(*) from user_account" +
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

func QueryUserAccount(userKey string, assetName string) ([]UserAccount, bool) {
	assetProperty := make([]UserAccount, 0)
	jsonMap := make(map[string]interface{})

	if len(userKey) > 0 {
		jsonMap["user_key"] = userKey
	}

	if len(assetName) > 0 {
		jsonMap["asset_name"] = assetName
	}

	pack, err := json.Marshal(jsonMap)
	if err != nil {
		return assetProperty, false
	}

	if assetProperty, ok := QueryUserAccountByJson(string(pack)); ok {
		return assetProperty, true
	}
	return assetProperty, false
}

func QueryUserAccountRow(userKey string, assetName string) (UserAccount, bool) {
	if assetProperty, ok := QueryUserAccount(userKey, assetName); ok {
		return assetProperty[0], true
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
			_, errMsg := CheckError(ErrorDataBase, err.Error())
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
