package mysqlpool

import (
	. "business_center/def"
	"encoding/json"
	"fmt"
	"time"
)

func QueryAllUserAddress() map[string]*UserAddress {
	mapUserAddress := make(map[string]*UserAddress)

	rows, err := db.Query("select b.user_key, b.user_name, b.user_class, c.id, c.name," +
		" c.full_name, a.address, a.private_key, " +
		" a.available_amount, a.frozen_amount, a.enabled, " +
		" unix_timestamp(a.create_time), unix_timestamp(a.update_time)" +
		" from user_address a " +
		" left join user_property b on a.user_key = b.user_key" +
		" left join asset_property c on a.asset_id = c.id;")
	if err != nil {
		fmt.Println(err.Error())
		return mapUserAddress
	}

	for rows.Next() {
		userAddress := &UserAddress{}
		rows.Scan(&userAddress.UserKey, &userAddress.UserName, &userAddress.UserClass, &userAddress.AssetID,
			&userAddress.AssetName, &userAddress.AssetFullName, &userAddress.Address, &userAddress.PrivateKey,
			&userAddress.AvailableAmount, &userAddress.FrozenAmount,
			&userAddress.Enabled, &userAddress.CreateTime, &userAddress.UpdateTime)
		mapUserAddress[userAddress.AssetName+"_"+userAddress.Address] = userAddress
	}
	return mapUserAddress
}

func QueryUserAddress(query string) string {
	var dat map[string]interface{}
	err := json.Unmarshal([]byte(query), &dat)
	if err != nil {
		return ""
	}

	sqlCount := "select count(*)" +
		" from user_address a " +
		" left join user_property b on a.user_key = b.user_key" +
		" left join asset_property c on a.asset_id = c.id"

	sqlQuery := "select b.user_key, b.user_name, b.user_class, c.id, c.name," +
		" c.full_name, a.address, a.private_key, " +
		" a.available_amount, a.frozen_amount, a.enabled, " +
		" unix_timestamp(a.create_time), unix_timestamp(a.update_time)" +
		" from user_address a " +
		" left join user_property b on a.user_key = b.user_key" +
		" left join asset_property c on a.asset_id = c.id"

	sqlWhere := ""
	resMap := make(map[string]interface{})

	params := make([]interface{}, 0)
	if len(dat) > 0 {
		if v, ok := dat["user_key"]; ok {
			if value, ok := v.(string); ok {
				if len(params) > 0 {
					sqlWhere += " and"
				} else {
					sqlWhere += " where"
				}
				sqlWhere += " b.user_key = ?"
				params = append(params, value)
			}
		}

		if v, ok := dat["user_name"]; ok {
			if value, ok := v.(string); ok {
				if len(params) > 0 {
					sqlWhere += " and"
				} else {
					sqlWhere += " where"
				}
				sqlWhere += " b.user_name = ?"
				params = append(params, value)
			}
		}

		if v, ok := dat["user_class"]; ok {
			if value, ok := v.(float64); ok {
				if len(params) > 0 {
					sqlWhere += " and"
				} else {
					sqlWhere += " where"
				}
				sqlWhere += " b.user_class = ?"
				params = append(params, int(value))
			}
		}

		if v, ok := dat["asset_name"]; ok {
			if value, ok := v.(string); ok {
				if len(params) > 0 {
					sqlWhere += " and"
				} else {
					sqlWhere += " where"
				}
				sqlWhere += " c.name = ?"
				params = append(params, value)
			}
		}

		if v, ok := dat["address"]; ok {
			if value, ok := v.(string); ok {
				if len(params) > 0 {
					sqlWhere += " and"
				} else {
					sqlWhere += " where"
				}
				sqlWhere += " a.address = ?"
				params = append(params, value)
			}
		}

		if v, ok := dat["max_amount"]; ok {
			if value, ok := v.(float64); ok {
				if len(params) > 0 {
					sqlWhere += " and"
				} else {
					sqlWhere += " where"
				}
				sqlWhere += " a.available_amount + a.frozen_amount <= ?"
				params = append(params, int64(value))
			}
		}

		if v, ok := dat["min_amount"]; ok {
			if value, ok := v.(float64); ok {
				if len(params) > 0 {
					sqlWhere += " and"
				} else {
					sqlWhere += " where"
				}
				sqlWhere += " a.available_amount + a.frozen_amount >= ?"
				params = append(params, int64(value))
			}
		}

		if v, ok := dat["create_time_begin"]; ok {
			if value, ok := v.(float64); ok {
				if len(params) > 0 {
					sqlWhere += " and"
				} else {
					sqlWhere += " where"
				}
				sqlWhere += " a.create_time >= ?"
				params = append(params, time.Unix(int64(value), 0).Format("2006-01-02 15:04:05"))
			}
		}

		if v, ok := dat["create_time_end"]; ok {
			if value, ok := v.(float64); ok {
				if len(params) > 0 {
					sqlWhere += " and"
				} else {
					sqlWhere += " where"
				}
				sqlWhere += " a.create_time <= ?"
				params = append(params, time.Unix(int64(value), 0).Format("2006-01-02 15:04:05"))
			}
		}

		row := db.QueryRow(sqlCount+sqlWhere, params...)
		if row != nil {
			var value int
			row.Scan(&value)
			resMap["total"] = value
		}

		if v, ok := dat["max_display"]; ok {
			if value, ok := v.(float64); ok {
				resMap["max_display"] = value
				sqlWhere += " limit ?, ?"

				pageIndex := 1
				if v, ok := dat["page_index"]; ok {
					if value, ok := v.(float64); ok {
						resMap["page_index"] = value
						pageIndex = int(value)
					}
				}
				params = append(params, (pageIndex-1)*int(value))
				params = append(params, value)
			}
		}
	}

	rows, err := db.Query(sqlQuery+sqlWhere, params...)
	if err != nil {
		return ""
	}

	var userAddresses []UserAddress
	for rows.Next() {
		var userAddress UserAddress
		rows.Scan(&userAddress.UserKey, &userAddress.UserName, &userAddress.UserClass, &userAddress.AssetID,
			&userAddress.AssetName, &userAddress.AssetFullName, &userAddress.Address, &userAddress.PrivateKey,
			&userAddress.AvailableAmount, &userAddress.FrozenAmount,
			&userAddress.Enabled, &userAddress.CreateTime, &userAddress.UpdateTime)
		userAddresses = append(userAddresses, userAddress)
	}
	resMap["address"] = userAddresses

	pack, err := json.Marshal(resMap)
	if err != nil {
		return ""
	}

	return string(pack)
}
