package mysqlpool

import (
	. "business_center/def"
	"fmt"
)

func QueryTransactionMessage(queryMap map[string]interface{}) ([]TransactionNotice, bool) {
	sqls := "select user_key, msg_id, trans_type, status, blockin_height, asset_name, address, amount," +
		" pay_fee, hash, order_id, unix_timestamp(time) from transaction_notice where true"

	messages := make([]TransactionNotice, 0)
	params := make([]interface{}, 0)

	if len(queryMap) > 0 {
		sqls += andConditions(queryMap, &params)
		sqls += andPagination(queryMap, &params)
	}
	sqls += " order by msg_id;"

	db := Get()
	rows, err := db.Query(sqls, params...)
	defer rows.Close()
	if err != nil {
		fmt.Println(err.Error())
		return messages, len(messages) > 0
	}

	var data TransactionNotice
	for rows.Next() {
		err := rows.Scan(&data.UserKey, &data.MsgID, &data.TransType, &data.Status, &data.BlockinHeight,
			&data.AssetName, &data.Address, &data.Amount, &data.PayFee, &data.Hash, &data.OrderID, &data.Time)
		if err == nil {
			messages = append(messages, data)
		}
	}
	return messages, len(messages) > 0
}
