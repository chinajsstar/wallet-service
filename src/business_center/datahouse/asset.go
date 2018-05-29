package datahouse

import (
	"api_router/base/data"
	"bastionpay_api/api/v1"
	"bastionpay_api/apibackend/v1/backend"
	"business_center/def"
	"business_center/mysqlpool"
	"encoding/json"
	"errors"
	l4g "github.com/alecthomas/log4go"
)

func SupportAssets(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	if userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey); !ok {
		res.Err, res.ErrMsg = def.CheckError(def.ErrorFailed, "无效用户-"+userProperty.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if assetProperty, ok := mysqlpool.QueryAssetProperty(nil); ok {
		supportAssetList := v1.AckSupportAssetList{}
		for _, value := range assetProperty {
			supportAssetList.Data = append(supportAssetList.Data, value.AssetName)
		}

		pack, err := json.Marshal(supportAssetList)
		if err != nil {
			res.Err, res.ErrMsg = def.CheckError(def.ErrorFailed, "返回数据包错误")
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		res.Value.Message = string(pack)
	}
	return nil
}

func AssetAttribute(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	if userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey); !ok {
		res.Err, res.ErrMsg = def.CheckError(def.ErrorFailed, "无效用户-"+userProperty.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqAssetsAttributeList{
		IsToken:      -1,
		TotalLines:   -1,
		PageIndex:    -1,
		MaxDispLines: -1,
	}

	if len(req.Argv.Message) > 0 {
		err := json.Unmarshal([]byte(req.Argv.Message), &params)
		if err != nil {
			res.Err, res.ErrMsg = def.CheckError(def.ErrorFailed, "解析Json失败-"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
	}

	queryMap := make(map[string]interface{})

	if len(params.AssetNames) > 0 {
		queryMap["asset_names"] = params.AssetNames
	}

	if params.IsToken > 0 {
		queryMap["is_token"] = params.IsToken
	}

	dataList := v1.AckAssetsAttributeList{
		PageIndex:    -1,
		MaxDispLines: -1,
		TotalLines:   -1,
	}

	if params.PageIndex > 0 {
		queryMap["page_index"] = params.PageIndex
		dataList.PageIndex = params.PageIndex
	}

	if params.MaxDispLines > 0 {
		queryMap["max_disp_lines"] = params.MaxDispLines
		dataList.MaxDispLines = params.MaxDispLines
	}

	if params.TotalLines == 0 {
		dataList.TotalLines = mysqlpool.QueryAssetPropertyCount(queryMap)
	} else if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	}

	if arr, ok := mysqlpool.QueryAssetProperty(queryMap); ok {
		for _, v := range arr {
			d := v1.AckAssetsAttribute{}
			d.AssetName = v.AssetName
			d.FullName = v.FullName
			d.IsToken = v.IsToken
			d.ParentName = v.ParentName
			d.DepositMin = v.DepositMin
			d.WithdrawalRate = v.WithdrawalRate
			d.WithdrawalValue = v.WithdrawalValue
			d.ConfirmationNum = v.ConfirmationNum
			d.Decimals = v.Decimals
			dataList.Data = append(dataList.Data, d)
		}
	}

	pack, err := json.Marshal(dataList)
	if err != nil {
		res.Err, res.ErrMsg = def.CheckError(def.ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func SpAssetAttribute(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	if userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey); !ok {
		res.Err, res.ErrMsg = def.CheckError(def.ErrorFailed, "无效用户-"+userProperty.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := backend.ReqSpAssetsAttributeList{
		IsToken:      -1,
		TotalLines:   -1,
		PageIndex:    -1,
		MaxDispLines: -1,
	}

	if len(req.Argv.Message) > 0 {
		err := json.Unmarshal([]byte(req.Argv.Message), &params)
		if err != nil {
			res.Err, res.ErrMsg = def.CheckError(def.ErrorFailed, "解析Json失败-"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
	}

	queryMap := make(map[string]interface{})

	if len(params.AssetNames) > 0 {
		queryMap["asset_names"] = params.AssetNames
	}

	if params.IsToken > 0 {
		queryMap["is_token"] = params.IsToken
	}

	dataList := backend.AckSpAssetsAttributeList{
		PageIndex:    -1,
		MaxDispLines: -1,
		TotalLines:   -1,
	}

	if params.PageIndex > 0 {
		queryMap["page_index"] = params.PageIndex
		dataList.PageIndex = params.PageIndex
	}

	if params.MaxDispLines > 0 {
		queryMap["max_disp_lines"] = params.MaxDispLines
		dataList.MaxDispLines = params.MaxDispLines
	}

	if params.TotalLines == 0 {
		dataList.TotalLines = mysqlpool.QueryAssetPropertyCount(queryMap)
	} else if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	}

	if arr, ok := mysqlpool.QueryAssetProperty(queryMap); ok {
		for _, v := range arr {
			d := backend.AckSpAssetsAttribute{}
			d.AssetName = v.AssetName
			d.FullName = v.FullName
			d.IsToken = v.IsToken
			d.ParentName = v.ParentName
			d.Logo = v.Logo
			d.DepositMin = v.DepositMin
			d.WithdrawalRate = v.WithdrawalRate
			d.WithdrawalValue = v.WithdrawalValue
			d.WithdrawalReserveRate = v.WithdrawalReserveRate
			d.WithdrawalAlertRate = v.WithdrawalAlertRate
			d.WithdrawalStategy = v.WithdrawalStategy
			d.ConfirmationNum = v.ConfirmationNum
			d.Decimals = v.Decimals
			d.GasFactor = v.GasFactor
			d.Debt = v.Debt
			d.ParkAmount = v.ParkAmount
			d.Enabled = v.Enabled
			dataList.Data = append(dataList.Data, d)
		}
	}

	pack, err := json.Marshal(dataList)
	if err != nil {
		res.Err, res.ErrMsg = def.CheckError(def.ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}
