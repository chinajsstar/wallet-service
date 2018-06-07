package data

import (
	"bastionpay_api/api/v1"
	"bastionpay_base/data"
	. "business/def"
	"business/mysqlpool"
	"encoding/json"
	"errors"
	l4g "github.com/alecthomas/log4go"
)

func HistoryTransactionMessage(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqTransactionMessage{
		MaxMessageID: -1,
		MinMessageID: -1,
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
	if userProperty.UserClass == 0 {
		queryMap["user_key"] = userProperty.UserKey
	}

	if params.MaxMessageID > 0 {
		queryMap["max_msg_id"] = params.MaxMessageID
	}

	if params.MinMessageID > 0 {
		queryMap["min_msg_id"] = params.MinMessageID
	}

	dataList := v1.AckTransactionMessageList{
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
		dataList.TotalLines = mysqlpool.QueryTransactionMessageCount(queryMap)
	}

	if arr, ok := mysqlpool.QueryTransactionMessage(queryMap); ok {
		for _, v := range arr {
			d := v1.AckTransactionMessage{}
			d.MsgID = v.MsgID
			d.TransType = v.TransType
			d.Status = v.Status
			d.BlockinHeight = v.BlockinHeight
			d.AssetName = v.AssetName
			d.Address = v.Address
			d.Amount = v.Amount
			d.PayFee = v.PayFee
			d.Balance = v.Balance
			d.Hash = v.Hash
			d.OrderID = v.OrderID
			d.UserOrderID = v.UserOrderID
			d.Time = v.Time
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
