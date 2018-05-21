package mysqlpool

import (
	. "business_center/def"
	"fmt"
)

func QueryTransactionOrder(queryMap map[string]interface{}) ([]TransactionOrder, bool) {
	sqls := "select id, user_key,trans_type,asset_name,address,amount,pay_fee,balance,hash,order_id,status,unix_timestamp(time)" +
		" from transaction_bill where true "

	orders := make([]TransactionOrder, 0)
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
		return orders, len(orders) > 0
	}

	var data TransactionOrder
	for rows.Next() {
		err := rows.Scan(&data.ID, &data.UserKey, &data.TransType, &data.AssetName, &data.Address, &data.Amount,
			&data.PayFee, &data.Balance, &data.Hash, &data.OrderID, &data.Status, &data.Time)
		if err == nil {
			orders = append(orders, data)
		}
	}
	return orders, len(orders) > 0
}

func QueryTransactionOrderCount(queryMap map[string]interface{}) int {
	sqls := "select count(*) from transaction_bill where true "

	count := 0
	params := make([]interface{}, 0)

	if len(queryMap) > 0 {
		sqls += andConditions(queryMap, &params)
	}

	db := Get()
	db.QueryRow(sqls, params...).Scan(&count)
	return count
}
