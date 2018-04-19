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
	d, err := sql.Open("mysql", "root:command@tcp(127.0.0.1:3306)/test?charset=utf8")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	db = d
	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	db.Ping()
}

func QueryAllUserProperty() (map[string]*UserProperty, error) {
	rows, err := db.Query("select user_key, user_name, user_class" +
		" from user_property;")
	if err != nil {
		return nil, err
	}

	mapUserProperty := make(map[string]*UserProperty)
	for rows.Next() {
		userProperty := &UserProperty{}
		rows.Scan(&userProperty.UserKey, &userProperty.UserName, &userProperty.UserClass)
		mapUserProperty[userProperty.UserKey] = userProperty
	}

	return mapUserProperty, nil
}

func QueryAllAssetProperty() (map[string]*AssetProperty, error) {
	rows, err := db.Query("select id, name, full_name, confirmation_num" +
		" from asset_property;")
	if err != nil {
		return nil, err
	}

	mapAssetProperty := make(map[string]*AssetProperty)
	for rows.Next() {
		assetProperty := &AssetProperty{}
		rows.Scan(&assetProperty.ID, &assetProperty.Name, &assetProperty.FullName, &assetProperty.ConfirmationNum)
		mapAssetProperty[assetProperty.Name] = assetProperty
	}

	return mapAssetProperty, nil
}
