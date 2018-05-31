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

func QueryAddress(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userProperty.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqUserAddress{
		MaxAllocationTime: -1,
		MinAllocationTime: -1,
		Address:           "",
		TotalLines:        -1,
		PageIndex:         -1,
		MaxDispLines:      -1,
	}

	err := json.Unmarshal([]byte(req.Argv.Message), &params)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	queryMap := make(map[string]interface{})
	if userProperty.UserClass == 0 {
		queryMap["user_key"] = userProperty.UserKey
	}

	if len(params.AssetNames) > 0 {
		queryMap["asset_names"] = params.AssetNames
	}

	if params.MaxAllocationTime > 0 {
		queryMap["max_allocation_time"] = params.MaxAllocationTime
	}

	if params.MinAllocationTime > 0 {
		queryMap["min_allocation_time"] = params.MinAllocationTime
	}

	if len(params.Address) > 0 {
		queryMap["address"] = params.Address
	}

	dataList := v1.AckUserAddressList{
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
		dataList.TotalLines = mysqlpool.QueryUserAddressCount(queryMap)
	}

	if arr, ok := mysqlpool.QueryUserAddress(queryMap); ok {
		for _, v := range arr {
			d := v1.AckUserAddress{}
			d.AssetName = v.AssetName
			d.Address = v.Address
			d.AllocationTime = v.AllocationTime

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

func SpQueryAddress(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userProperty.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := backend.SpReqUserAddress{
		UserClass:         -1,
		MaxAllocationTime: -1,
		MinAllocationTime: -1,
		Address:           "",
		TotalLines:        -1,
		PageIndex:         -1,
		MaxDispLines:      -1,
	}

	err := json.Unmarshal([]byte(req.Argv.Message), &params)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	queryMap := make(map[string]interface{})
	if userProperty.UserClass == 0 {
		queryMap["user_key"] = userProperty.UserKey
	} else {
		queryMap["user_key"] = params.UserKey
	}

	if params.UserClass >= 0 {
		queryMap["user_class"] = params.UserClass
	}

	if len(params.AssetNames) > 0 {
		queryMap["asset_names"] = params.AssetNames
	}

	if params.MaxAllocationTime > 0 {
		queryMap["max_allocation_time"] = params.MaxAllocationTime
	}

	if params.MinAllocationTime > 0 {
		queryMap["min_allocation_time"] = params.MinAllocationTime
	}

	if len(params.Address) > 0 {
		queryMap["address"] = params.Address
	}

	dataList := backend.SpAckUserAddressList{
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
		dataList.TotalLines = mysqlpool.QueryUserAddressCount(queryMap)
	}

	if arr, ok := mysqlpool.QueryUserAddress(queryMap); ok {
		for _, v := range arr {
			d := backend.SpAckUserAddress{}
			d.UserKey = v.UserKey
			d.UserClass = v.UserClass
			d.AssetName = v.AssetName
			d.Address = v.Address
			d.PrivateKey = v.PrivateKey
			d.AvailableAmount = v.AvailableAmount
			d.FrozenAmount = v.FrozenAmount
			d.Enabled = v.Enabled
			d.CreateTime = v.CreateTime
			d.AllocationTime = v.AllocationTime
			d.UpdateTime = v.UpdateTime
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
