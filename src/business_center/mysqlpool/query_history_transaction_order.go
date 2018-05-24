package mysqlpool

import (
	. "business_center/def"
	"fmt"
)

func QueryTransactionBill(queryMap map[string]interface{}) ([]TransactionBill, bool) {
	sqls := "select id,user_key,order_id,user_order_id,trans_type,asset_name,address,amount,pay_fee,miner_fee,balance,hash,status," +
		" blockin_height,ifnull(unix_timestamp(create_order_time),0),ifnull(unix_timestamp(blockin_time),0),ifnull(unix_timestamp(confirm_time),0)" +
		" from transaction_bill_view where true "

	dataList := make([]TransactionBill, 0)
	params := make([]interface{}, 0)

	if len(queryMap) > 0 {
		sqls += andConditions(queryMap, &params)
		sqls += " order by id "
		sqls += andPagination(queryMap, &params)
	}

	db := Get()
	rows, err := db.Query(sqls, params...)
	defer rows.Close()
	if err != nil {
		fmt.Println(err.Error())
		return dataList, len(dataList) > 0
	}

	var data TransactionBill
	for rows.Next() {
		err := rows.Scan(&data.ID, &data.UserKey, &data.OrderID, &data.UserOrderID, &data.TransType, &data.AssetName, &data.Address,
			&data.Amount, &data.PayFee, &data.MinerFee, &data.Balance, &data.Hash, &data.Status, &data.BlockinHeight,
			&data.CreateOrderTime, &data.BlockinTime, &data.ConfirmTime)
		if err == nil {
			dataList = append(dataList, data)
		} else {
			fmt.Println(err.Error())
		}
	}
	return dataList, len(dataList) > 0
}

func QueryTransactionBillDaily(queryMap map[string]interface{}) ([]TransactionBillDaily, bool) {
	sqls := "select period, user_key, asset_name, sum_dp_amount, sum_wd_amount, sum_pay_fee, sum_miner_fee, " +
		" pre_balance, last_balance from transaction_bill_daily where true "

	dataList := make([]TransactionBillDaily, 0)
	params := make([]interface{}, 0)

	if len(queryMap) > 0 {
		sqls += andConditions(queryMap, &params)
		sqls += " order by period "
		sqls += andPagination(queryMap, &params)
	}

	db := Get()
	rows, err := db.Query(sqls, params...)
	defer rows.Close()
	if err != nil {
		fmt.Println(err.Error())
		return dataList, len(dataList) > 0
	}

	var data TransactionBillDaily
	for rows.Next() {
		err := rows.Scan(&data.Period, &data.UserKey, &data.AssetName, &data.SumDPAmount, &data.SumWDAmount,
			&data.SumPayFee, &data.SumMinerFee, &data.PreBalance, &data.LastBalance)
		if err == nil {
			dataList = append(dataList, data)
		}
	}
	return dataList, len(dataList) > 0
}

func QueryTransactionBillCount(queryMap map[string]interface{}) int {
	sqls := "select count(*) from transaction_bill_view where true "

	count := 0
	params := make([]interface{}, 0)

	if len(queryMap) > 0 {
		sqls += andConditions(queryMap, &params)
	}

	db := Get()
	db.QueryRow(sqls, params...).Scan(&count)
	return count
}

func QueryTransactionBillDailyCount(queryMap map[string]interface{}) int {
	sqls := "select count(*) from transaction_bill_daily where true "

	count := 0
	params := make([]interface{}, 0)

	if len(queryMap) > 0 {
		sqls += andConditions(queryMap, &params)
	}

	db := Get()
	db.QueryRow(sqls, params...).Scan(&count)
	return count
}
