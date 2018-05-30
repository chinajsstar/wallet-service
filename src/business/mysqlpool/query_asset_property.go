package mysqlpool

import (
	. "business/def"
	"fmt"
)

func QueryAssetProperty(queryMap map[string]interface{}) ([]AssetProperty, bool) {
	sqls := "select asset_name,full_name,is_token,parent_name,logo,deposit_min,withdrawal_rate," +
		"withdrawal_value,withdrawal_reserve_rate,withdrawal_alert_rate,withdrawal_stategy,confirmation_num," +
		"decimals,gas_factor,debt,park_amount from asset_property where true"

	assetProperty := make([]AssetProperty, 0)
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
		return assetProperty, len(assetProperty) > 0
	}

	var data AssetProperty
	for rows.Next() {
		err := rows.Scan(&data.AssetName, &data.FullName, &data.IsToken, &data.ParentName, &data.Logo,
			&data.DepositMin, &data.WithdrawalRate, &data.WithdrawalValue, &data.WithdrawalReserveRate,
			&data.WithdrawalAlertRate, &data.WithdrawalStategy, &data.ConfirmationNum, &data.Decimals,
			&data.GasFactor, &data.Debt, &data.ParkAmount)
		if err == nil {
			assetProperty = append(assetProperty, data)
		}
	}
	return assetProperty, len(assetProperty) > 0
}

func QueryAssetPropertyCount(queryMap map[string]interface{}) int {
	sqls := "select count(*) from asset_property" +
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

func QueryAssetPropertyByName(assetName string) (AssetProperty, bool) {
	queryMap := make(map[string]interface{})
	queryMap["asset_name"] = assetName
	if assetProperty, ok := QueryAssetProperty(queryMap); ok {
		return assetProperty[0], true
	}
	return AssetProperty{}, false
}

func SetAssetProperty(assetProperty *AssetProperty) error {
	db := Get()
	params := make([]interface{}, 0)
	params = append(params, assetProperty.FullName, assetProperty.IsToken, assetProperty.ParentName, assetProperty.Logo,
		assetProperty.DepositMin, assetProperty.WithdrawalRate, assetProperty.WithdrawalValue,
		assetProperty.WithdrawalReserveRate, assetProperty.WithdrawalAlertRate, assetProperty.WithdrawalStategy,
		assetProperty.ConfirmationNum, assetProperty.Decimals, assetProperty.GasFactor, assetProperty.Debt,
		assetProperty.ParkAmount, assetProperty.Enabled, assetProperty.AssetName)

	ret, err := db.Exec("update asset_property set full_name = ?, is_token = ?, parent_name = ?, logo = ?,"+
		" deposit_min = ?,withdrawal_rate = ?, withdrawal_value = ?, withdrawal_reserve_rate = ?,"+
		" withdrawal_alert_rate = ?, withdrawal_stategy = ?, confirmation_num = ?, decimals = ?, gas_factor = ?,"+
		" debt = ?, park_amount = ?, enabled = ? where asset_name = ?", params...)
	if err != nil {
		return err
	}

	rowAffected, err := ret.RowsAffected()
	if err != nil {
		return err
	}

	if rowAffected > 0 {
		return nil
	}

	_, err = db.Exec("insert asset_property (full_name, is_token, parent_name, logo, deposit_min, withdrawal_rate,"+
		" withdrawal_value,withdrawal_reserve_rate, withdrawal_alert_rate, withdrawal_stategy, confirmation_num,"+
		" decimals, gas_factor, debt, park_amount, asset_name, enabled)"+
		" values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", params...)
	return err
}
