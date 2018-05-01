package mysqlpool

import (
	. "business_center/def"
	"encoding/json"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"time"
)

func QueryUserAccount(query string) ([]UserAccount, bool) {
	sqls := "select user_key,user_class,asset_id,available_amount,frozen_amount," +
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
		err := rows.Scan(&data.UserKey, &data.UserClass, &data.AssetID, &data.AvailableAmount,
			&data.FrozenAmount, &data.CreateTime, &data.UpdateTime)
		if err == nil {
			userAccount = append(userAccount, data)
		}
	}
	return userAccount, len(userAccount) > 0
}

func QueryUserAccountCount(query string) int {
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

func QueryUserAccountByKey(userKey string) (UserAccount, bool) {
	query := fmt.Sprintf("{\"user_key\":\"%s\"}", userKey)
	if assetProperty, ok := QueryUserAccount(query); ok {
		return assetProperty[0], true
	}
	return UserAccount{}, false
}

func AddUserAccount(userKey string, userClass int, assetID int) error {
	db := Get()
	nowTM := time.Now().UTC().Format(TimeFormat)
	_, err := db.Exec("insert user_account (user_key, user_class, asset_id, available_amount, frozen_amount,"+
		" create_time, update_time) values (?, ?, ?, 0, 0, ?, ?);",
		userKey, userClass, assetID, nowTM, nowTM)
	if err != nil {
		count := 0
		row := db.QueryRow("select count(*) from user_account where user_key = ? and asset_id = ?;", userKey, assetID)
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
