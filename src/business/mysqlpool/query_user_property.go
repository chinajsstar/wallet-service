package mysqlpool

import (
	. "business/def"
	"fmt"
)

func QueryUserProperty(queryMap map[string]interface{}) ([]UserProperty, bool) {
	sqls := "select user_key,user_class,is_frozen,unix_timestamp(create_time),unix_timestamp(update_time)" +
		" from user_property where true "

	userProperty := make([]UserProperty, 0)
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
		return userProperty, len(userProperty) > 0
	}

	var data UserProperty
	for rows.Next() {
		err := rows.Scan(&data.UserKey, &data.UserClass, &data.IsFrozen, &data.CreateTime, &data.UpdateTime)
		if err == nil {
			userProperty = append(userProperty, data)
		}
	}
	return userProperty, len(userProperty) > 0
}

func QueryUserPropertyCount(queryMap map[string]interface{}) int {
	sqls := "select count(*) from user_property" +
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

func QueryUserPropertyByKey(userKey string) (UserProperty, bool) {
	queryMap := make(map[string]interface{})
	queryMap["user_key"] = userKey
	if userProperty, ok := QueryUserProperty(queryMap); ok {
		return userProperty[0], true
	}
	return UserProperty{}, false
}
