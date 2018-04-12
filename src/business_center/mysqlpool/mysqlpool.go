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
	rows, err := db.Query("select user_id, user_name, user_class from user_property;")
	if err != nil {
		return nil, err
	}

	mapUserProperty := make(map[string]*UserProperty)
	for rows.Next() {
		userProperty := &UserProperty{}
		rows.Scan(&userProperty.UserID, &userProperty.UserName, &userProperty.UserClass)
		mapUserProperty[userProperty.UserID] = userProperty
	}

	return mapUserProperty, nil
}

func QueryAllAssetProperty() (map[string]*AssetProperty, error) {
	rows, err := db.Query("select id, name, full_name, confirmation_num from asset_property;")
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

func QueryAllUserAddress() (map[string]*UserAddress, error) {
	sql := "select a.user_id, b.user_class, a.asset_id, c.name as asset_name, a.address, a.private_key, "
	sql += "a.available_amount, a.frozen_amount, a.enabled, a.create_time, a.update_time from user_address a "
	sql += "left join user_property b on a.user_id = b.user_id "
	sql += "left join asset_property c on a.asset_id = c.id;"

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}

	mapUserAddress := make(map[string]*UserAddress)
	for rows.Next() {
		userAddress := &UserAddress{}
		rows.Scan(&userAddress.UserID, &userAddress.UserClass, &userAddress.AssetID, &userAddress.AssetName,
			&userAddress.Address, &userAddress.PrivateKey, &userAddress.AvailableAmount, &userAddress.FrozenAmount,
			&userAddress.Enabled, &userAddress.CreateTime, &userAddress.UpdateTime)
		mapUserAddress[userAddress.AssetName+"_"+userAddress.Address] = userAddress
	}

	return mapUserAddress, nil
}
