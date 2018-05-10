package mysqlpool

import (
	. "business_center/def"
	"fmt"
)

func QueryTransactionOrder(queryMap map[string]interface{}) ([]TransactionOrder, bool) {
	sqls := "select user_key,trans_type,asset_name,amount,pay_fee,hash,order_id,status,unix_timestamp(time)" +
		" from transaction_order_view where true"

	orders := make([]TransactionOrder, 0)
	params := make([]interface{}, 0)

	if len(queryMap) > 0 {
		sqls += andConditions(queryMap, &params)
		sqls += andPagination(queryMap, &params)
	}

	db := Get()
	rows, err := db.Query(sqls, params...)
	if err != nil {
		fmt.Println(err.Error())
		return orders, len(orders) > 0
	}

	var data TransactionOrder
	for rows.Next() {
		err := rows.Scan(&data.UserKey, &data.TransType, &data.AssetName, &data.Amount, &data.PayFee,
			&data.Hash, &data.OrderID, &data.Status, &data.Time)
		if err == nil {
			orders = append(orders, data)
		}
	}
	return orders, len(orders) > 0
}