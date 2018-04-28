package mysqlpool

import (
	. "business_center/def"
	"encoding/json"
	"fmt"
)

func QueryUserProperty(query string) ([]UserProperty, bool) {
	sqls := "select user_key,user_name,user_class,phone,email,salt,password,google_auth,public_key," +
		"callback_url,level,last_login_time,last_login_ip,last_login_mac,unix_timestamp(create_time)," +
		"unix_timestamp(update_time),is_frozen,time_zone,country,language from user_property" +
		" where true"

	userProperty := make([]UserProperty, 0)
	params := make([]interface{}, 0)

	if len(query) > 0 {
		var queryMap map[string]interface{}
		err := json.Unmarshal([]byte(query), &queryMap)
		if err != nil {
			return userProperty, len(userProperty) > 0
		}

		sqls += andConditions(queryMap, &params)
		sqls += andPagination(queryMap, &params)
	}

	db := Get()
	rows, err := db.Query(sqls, params...)
	if err != nil {
		fmt.Println(err.Error())
		return userProperty, len(userProperty) > 0
	}

	var data UserProperty
	for rows.Next() {
		err := rows.Scan(&data.UserKey, &data.UserName, &data.UserClass, &data.Phone,
			&data.Email, &data.Salt, &data.Password, &data.GoogleAuth,
			&data.PublicKey, &data.CallbackUrl, &data.Level, &data.LastLoginTime,
			&data.LastLoginIp, &data.LastLoginMac, &data.CreateTime, &data.UpdateTime,
			&data.IsFrozen, &data.TimeZone, &data.Conutry, &data.Language)
		if err == nil {
			userProperty = append(userProperty, data)
		}
	}
	return userProperty, len(userProperty) > 0
}

func QueryUserPropertyCount(query string) int {
	sqls := "select count(*) from user_property" +
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

func QueryUserPropertyByKey(userKey string) (UserProperty, bool) {
	query := fmt.Sprintf("{\"user_key\":\"%s\"}", userKey)
	if userProperty, ok := QueryUserProperty(query); ok {
		return userProperty[0], true
	}
	return UserProperty{}, false
}
