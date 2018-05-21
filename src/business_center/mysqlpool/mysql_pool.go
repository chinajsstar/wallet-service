package mysqlpool

import (
	"api_router/base/config"
	. "business_center/def"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"math"
	"time"
)

var db *sql.DB = nil

func Get() *sql.DB {
	return db
}

func init() {
	var dataSourceName string
	err := config.LoadJsonNode(config.GetBastionPayConfigDir()+"/cobank.json", "db", &dataSourceName)
	if err != nil {
		return
	}

	d, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	db = d
	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	db.Ping()

	UpdatePayTokenAddress()
}

func andConditions(queryMap map[string]interface{}, params *[]interface{}) string {
	sqls := ""
	for key, value := range queryMap {
		switch key {
		case "id":
			if value, ok := value.(int64); ok {
				sqls += " and id = ?"
				*params = append(*params, value)
			}
		case "user_key":
			if value, ok := value.(string); ok {
				sqls += " and user_key = ?"
				*params = append(*params, value)
			}
		case "user_class":
			if value, ok := value.(string); ok {
				sqls += " and user_class = ?"
				*params = append(*params, value)
			}
		case "asset_name":
			if value, ok := value.(string); ok {
				sqls += " and asset_name = ?"
				*params = append(*params, value)
			}
		case "asset_names":
			if value, ok := value.([]string); ok {
				sqls += " and asset_name in (true"
				for _, value := range value {
					sqls += ", ?"
					*params = append(*params, value)
				}
				sqls += ")"
			}
		case "is_token":
			if value, ok := value.(float64); ok {
				sqls += " and is_token = ?"
				*params = append(*params, value)
			}
		case "address":
			if value, ok := value.(string); ok {
				sqls += " and address = ?"
				*params = append(*params, value)
			}
		case "trans_type":
			if value, ok := value.(float64); ok {
				sqls += " and trans_type = ?"
				*params = append(*params, int(value))
			}
		case "status":
			if value, ok := value.(float64); ok {
				sqls += " and status = ?"
				*params = append(*params, int(value))
			}
		case "max_amount":
			if value, ok := value.(float64); ok {
				sqls += " and amount <= ?"
				*params = append(*params, int64(value))
			}
		case "min_amount":
			if value, ok := value.(float64); ok {
				sqls += " and amount >= ?"
				*params = append(*params, int64(value))
			}
		case "max_create_time":
			if value, ok := value.(float64); ok {
				sqls += " and create_time <= ?"
				*params = append(*params, time.Unix(int64(value), 0).Format(TimeFormat))
			}
		case "min_create_time":
			if value, ok := value.(float64); ok {
				sqls += " and create_time >= ?"
				*params = append(*params, time.Unix(int64(value), 0).Format(TimeFormat))
			}
		case "max_update_time":
			if value, ok := value.(float64); ok {
				sqls += " and update_time <= ?"
				*params = append(*params, time.Unix(int64(value), 0).Format(TimeFormat))
			}
		case "min_update_time":
			if value, ok := value.(float64); ok {
				sqls += " and update_time >= ?"
				*params = append(*params, time.Unix(int64(value), 0).Format(TimeFormat))
			}
		case "max_msg_id":
			if value, ok := value.(float64); ok {
				sqls += " and msg_id <= ?"
				*params = append(*params, int(value))
			}
		case "min_msg_id":
			if value, ok := value.(float64); ok {
				sqls += " and msg_id >= ?"
				*params = append(*params, int(value))
			}
		}
	}
	return sqls
}

func andPagination(queryMap map[string]interface{}, params *[]interface{}) string {
	sqls := ""
	if value, ok := queryMap["max_disp_lines"]; ok {
		if value, ok := value.(float64); ok {
			sqls += " limit ?, ?;"

			var pageIndex float64 = 1
			if v, ok := queryMap["page_index"]; ok {
				if value, ok := v.(float64); ok {
					pageIndex = math.Max(pageIndex, value)
				}
			}
			*params = append(*params, (int64(pageIndex)-1)*int64(value))
			*params = append(*params, value)
		}
	}
	return sqls
}
