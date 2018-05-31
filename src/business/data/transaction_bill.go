package data

import (
	"api_router/base/data"
	"bastionpay_api/api/v1"
	"bastionpay_api/apibackend/v1/backend"
	. "business/def"
	"business/mysqlpool"
	"encoding/json"
	"errors"
	l4g "github.com/alecthomas/log4go"
)

func HistoryTransactionBill(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqTransactionBill{
		ID:             -1,
		OrderID:        "",
		AssetName:      "",
		Address:        "",
		TransType:      -1,
		Status:         -1,
		Hash:           "",
		MaxAmount:      -1,
		MinAmount:      -1,
		MaxConfirmTime: -1,
		MinConfirmTime: -1,
		TotalLines:     -1,
		PageIndex:      -1,
		MaxDispLines:   -1,
	}

	if len(req.Argv.Message) > 0 {
		err := json.Unmarshal([]byte(req.Argv.Message), &params)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "解析Json失败-"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
	}

	queryMap := make(map[string]interface{})
	if userProperty.UserClass == 0 {
		queryMap["user_key"] = userProperty.UserKey
	}

	if params.ID > 0 {
		queryMap["id"] = params.ID
	}

	if len(params.OrderID) > 0 {
		queryMap["order_id"] = params.OrderID
	}

	if len(params.AssetName) > 0 {
		queryMap["asset_name"] = params.AssetName
	}

	if len(params.Address) > 0 {
		queryMap["address"] = params.Address
	}

	if params.TransType >= 0 {
		queryMap["trans_type"] = params.TransType
	}

	if params.Status >= 0 {
		queryMap["status"] = params.Status
	}

	if len(params.Hash) > 0 {
		queryMap["hash"] = params.Hash
	}

	if params.MaxAmount >= 0 {
		queryMap["max_amount"] = params.MaxAmount
	}

	if params.MinAmount >= 0 {
		queryMap["min_amount"] = params.MinAmount
	}

	if params.MaxConfirmTime > 0 {
		queryMap["max_confirm_time"] = params.MaxConfirmTime
	}

	if params.MinConfirmTime > 0 {
		queryMap["min_confirm_time"] = params.MinConfirmTime
	}

	dataList := v1.AckTransactionBillList{
		TotalLines:   -1,
		PageIndex:    -1,
		MaxDispLines: -1,
	}

	if params.PageIndex > 0 {
		queryMap["page_index"] = params.PageIndex
		dataList.PageIndex = params.PageIndex
	}

	if params.MaxDispLines > 0 {
		queryMap["max_disp_lines"] = params.MaxDispLines
		dataList.MaxDispLines = params.MaxDispLines
	}

	if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	} else {
		dataList.TotalLines = mysqlpool.QueryTransactionBillCount(queryMap)
	}

	if arr, ok := mysqlpool.QueryTransactionBill(queryMap); ok {
		for _, v := range arr {
			d := v1.AckTransactionBill{}
			d.ID = v.ID
			d.OrderID = v.OrderID
			d.UserOrderID = v.UserOrderID
			d.AssetName = v.AssetName
			d.Address = v.Address
			d.TransType = v.TransType
			d.Amount = v.Amount
			d.PayFee = v.PayFee
			d.Balance = v.Balance
			d.Hash = v.Hash
			d.Status = v.Status
			d.BlockinHeight = v.BlockinHeight
			d.CreateOrderTime = v.CreateOrderTime
			d.BlockinTime = v.BlockinTime
			d.ConfirmTime = v.ConfirmTime
			dataList.Data = append(dataList.Data, d)
		}
	}

	pack, err := json.Marshal(dataList)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func HistoryTransactionBillDaily(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userProperty.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqTransactionBillDaily{
		AssetName:    "",
		MaxPeriod:    -1,
		MinPeriod:    -1,
		TotalLines:   -1,
		PageIndex:    -1,
		MaxDispLines: -1,
	}

	if len(req.Argv.Message) > 0 {
		err := json.Unmarshal([]byte(req.Argv.Message), &params)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "解析Json失败-"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
	}

	queryMap := make(map[string]interface{})
	queryMap["user_key"] = userProperty.UserKey

	if len(params.AssetName) > 0 {
		queryMap["asset_name"] = params.AssetName
	}

	if params.MaxPeriod >= 0 {
		queryMap["max_period"] = params.MaxPeriod
	}

	if params.MinPeriod >= 0 {
		queryMap["min_period"] = params.MinPeriod
	}

	dataList := v1.AckTransactionBillDailyList{
		TotalLines:   -1,
		PageIndex:    -1,
		MaxDispLines: -1,
	}

	if params.PageIndex > 0 {
		queryMap["page_index"] = params.PageIndex
		dataList.PageIndex = params.PageIndex
	}

	if params.MaxDispLines > 0 {
		queryMap["max_disp_lines"] = params.MaxDispLines
		dataList.MaxDispLines = params.MaxDispLines
	}

	if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	} else {
		dataList.TotalLines = mysqlpool.QueryTransactionBillDailyCount(queryMap)
	}

	if arr, ok := mysqlpool.QueryTransactionBillDaily(queryMap); ok {
		for _, v := range arr {
			d := v1.AckTransactionBillDaily{}
			d.Period = v.Period
			d.AssetName = v.AssetName
			d.SumDPAmount = v.SumDPAmount
			d.SumWDAmount = v.SumWDAmount
			d.SumPayFee = v.SumPayFee
			d.PreBalance = v.PreBalance
			d.LastBalance = v.LastBalance
			dataList.Data = append(dataList.Data, d)
		}
	}

	pack, err := json.Marshal(dataList)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func SpHistoryTransactionBill(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userPropertyMap := mysqlpool.QueryUserPropertyMap(nil)
	userProperty, ok := userPropertyMap[userKey]
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := backend.SpReqTransactionBill{
		ID:             -1,
		UserKey:		"",
		OrderID:        "",
		AssetName:      "",
		Address:        "",
		TransType:      -1,
		Status:         -1,
		Hash:           "",
		MaxAmount:      -1,
		MinAmount:      -1,
		MaxConfirmTime: -1,
		MinConfirmTime: -1,
		TotalLines:     -1,
		PageIndex:      -1,
		MaxDispLines:   -1,
	}

	if len(req.Argv.Message) > 0 {
		err := json.Unmarshal([]byte(req.Argv.Message), &params)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "解析Json失败-"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
	}

	queryMap := make(map[string]interface{})
	if len(params.UserKey) > 0 {
		queryMap["user_key"] = userProperty.UserKey
	}

	if params.ID > 0 {
		queryMap["id"] = params.ID
	}

	if len(params.OrderID) > 0 {
		queryMap["order_id"] = params.OrderID
	}

	if len(params.AssetName) > 0 {
		queryMap["asset_name"] = params.AssetName
	}

	if len(params.Address) > 0 {
		queryMap["address"] = params.Address
	}

	if params.TransType >= 0 {
		queryMap["trans_type"] = params.TransType
	}

	if params.Status >= 0 {
		queryMap["status"] = params.Status
	}

	if len(params.Hash) > 0 {
		queryMap["hash"] = params.Hash
	}

	if params.MaxAmount >= 0 {
		queryMap["max_amount"] = params.MaxAmount
	}

	if params.MinAmount >= 0 {
		queryMap["min_amount"] = params.MinAmount
	}

	if params.MaxConfirmTime > 0 {
		queryMap["max_confirm_time"] = params.MaxConfirmTime
	}

	if params.MinConfirmTime > 0 {
		queryMap["min_confirm_time"] = params.MinConfirmTime
	}

	dataList := backend.SpAckTransactionBillList{
		TotalLines:   -1,
		PageIndex:    -1,
		MaxDispLines: -1,
	}

	if params.PageIndex > 0 {
		queryMap["page_index"] = params.PageIndex
		dataList.PageIndex = params.PageIndex
	}

	if params.MaxDispLines > 0 {
		queryMap["max_disp_lines"] = params.MaxDispLines
		dataList.MaxDispLines = params.MaxDispLines
	}

	if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	} else {
		dataList.TotalLines = mysqlpool.QueryTransactionBillCount(queryMap)
	}

	if arr, ok := mysqlpool.QueryTransactionBill(queryMap); ok {
		for _, v := range arr {
			d := backend.SpAckTransactionBill{}
			d.UserKey = v.UserKey
			if u, ok := userPropertyMap[v.UserKey]; ok {
				d.UserName = u.UserName
			}
			d.ID = v.ID
			d.OrderID = v.OrderID
			d.UserOrderID = v.UserOrderID
			d.AssetName = v.AssetName
			d.Address = v.Address
			d.TransType = v.TransType
			d.Amount = v.Amount
			d.PayFee = v.PayFee
			d.MinerFee = v.MinerFee
			d.Balance = v.Balance
			d.Hash = v.Hash
			d.Status = v.Status
			d.BlockinHeight = v.BlockinHeight
			d.CreateOrderTime = v.CreateOrderTime
			d.BlockinTime = v.BlockinTime
			d.ConfirmTime = v.ConfirmTime
			dataList.Data = append(dataList.Data, d)
		}
	}

	pack, err := json.Marshal(dataList)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func SpHistoryTransactionBillDaily(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userPropertyMap := mysqlpool.QueryUserPropertyMap(nil)
	userProperty, ok := userPropertyMap[userKey]
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := backend.SpReqTransactionBillDaily{
		UserKey:      "",
		AssetName:    "",
		MaxPeriod:    -1,
		MinPeriod:    -1,
		TotalLines:   -1,
		PageIndex:    -1,
		MaxDispLines: -1,
	}

	if len(req.Argv.Message) > 0 {
		err := json.Unmarshal([]byte(req.Argv.Message), &params)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "解析Json失败-"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
	}

	queryMap := make(map[string]interface{})

	if len(params.UserKey) > 0 {
		queryMap["user_key"] = userProperty.UserKey
	}

	if len(params.AssetName) > 0 {
		queryMap["asset_name"] = params.AssetName
	}

	if params.MaxPeriod >= 0 {
		queryMap["max_period"] = params.MaxPeriod
	}

	if params.MinPeriod >= 0 {
		queryMap["min_period"] = params.MinPeriod
	}

	dataList := backend.SpAckTransactionBillDailyList{
		TotalLines:   -1,
		PageIndex:    -1,
		MaxDispLines: -1,
	}

	if params.PageIndex > 0 {
		queryMap["page_index"] = params.PageIndex
		dataList.PageIndex = params.PageIndex
	}

	if params.MaxDispLines > 0 {
		queryMap["max_disp_lines"] = params.MaxDispLines
		dataList.MaxDispLines = params.MaxDispLines
	}

	if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	} else {
		dataList.TotalLines = mysqlpool.QueryTransactionBillDailyCount(queryMap)
	}

	if arr, ok := mysqlpool.QueryTransactionBillDaily(queryMap); ok {
		for _, v := range arr {
			d := backend.SpAckTransactionBillDaily{}
			d.UserKey = v.UserKey
			if u, ok := userPropertyMap[v.UserKey]; ok {
				d.UserName = u.UserName
			}
			d.Period = v.Period
			d.AssetName = v.AssetName
			d.SumDPAmount = v.SumDPAmount
			d.SumWDAmount = v.SumWDAmount
			d.SumPayFee = v.SumPayFee
			d.SumMinerFee = v.SumMinerFee
			d.PreBalance = v.PreBalance
			d.LastBalance = v.LastBalance
			dataList.Data = append(dataList.Data, d)
		}
	}

	pack, err := json.Marshal(dataList)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}
