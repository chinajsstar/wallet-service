package mysqlpool

import (
	. "business_center/def"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB = nil

func Get() *sql.DB {
	return db
}

func init() {
	d, err := sql.Open("mysql", "root:root@tcp(192.168.21.109:3306)/wallet?charset=utf8")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	db = d
	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	db.Ping()
}

func QueryAllUserProperty() map[string]*UserProperty {
	mapUserProperty := make(map[string]*UserProperty)

	rows, err := db.Query("select user_key, user_name, user_class" +
		" from user_property;")
	if err != nil {
		fmt.Println(err.Error())
		return mapUserProperty
	}

	for rows.Next() {
		userProperty := &UserProperty{}
		rows.Scan(&userProperty.UserKey, &userProperty.UserName, &userProperty.UserClass)
		mapUserProperty[userProperty.UserKey] = userProperty
	}
	return mapUserProperty
}

func QueryAllAssetProperty() map[string]*AssetProperty {
	mapAssetProperty := make(map[string]*AssetProperty)

	rows, err := db.Query("select id, name, full_name, is_token, coin_name, confirmation_num" +
		" from asset_property;")
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	for rows.Next() {
		assetProperty := &AssetProperty{}
		rows.Scan(&assetProperty.ID, &assetProperty.Name, &assetProperty.FullName, &assetProperty.IsToken, &assetProperty.CoinName,
			&assetProperty.ConfirmationNum)
		mapAssetProperty[assetProperty.Name] = assetProperty
	}
	return mapAssetProperty
}
