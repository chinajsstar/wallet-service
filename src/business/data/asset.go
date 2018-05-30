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

func SupportAssets(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	if userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey); !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userProperty.UserKey)
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
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
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
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userProperty.UserKey)
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
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "解析Json失败-"+err.Error())
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

	if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	} else {
		dataList.TotalLines = mysqlpool.QueryAssetPropertyCount(queryMap)
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
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func SpGetAssetAttribute(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	if userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey); !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userProperty.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := backend.SpReqAssetsAttributeList{
		IsToken:      -1,
		Enabled:      -1,
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

	if len(params.AssetNames) > 0 {
		queryMap["asset_names"] = params.AssetNames
	}

	if params.IsToken >= 0 {
		queryMap["is_token"] = params.IsToken
	}

	if params.Enabled >= 0 {
		queryMap["enabled"] = params.Enabled
	}

	dataList := backend.SpAckAssetsAttributeList{
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

	if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	} else {
		dataList.TotalLines = mysqlpool.QueryAssetPropertyCount(queryMap)
	}

	if arr, ok := mysqlpool.QueryAssetProperty(queryMap); ok {
		for _, v := range arr {
			d := backend.SpAckAssetsAttribute{}
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
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func SpSetAssetAttribute(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userProperty.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if !(userProperty.UserClass == 1 || userProperty.UserClass == 2) {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "该用户不能执行该操作")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := backend.SpReqSetAssetAttribute{
		AssetName:             "",
		FullName:              "",
		IsToken:               -1,
		ParentName:            "",
		Logo:                  "",
		DepositMin:            -1,
		WithdrawalRate:        -1,
		WithdrawalValue:       -1,
		WithdrawalReserveRate: -1,
		WithdrawalAlertRate:   -1,
		WithdrawalStategy:     -1,
		ConfirmationNum:       -1,
		Decimals:              -1,
		GasFactor:             -1,
		Debt:                  -1,
		ParkAmount:            -1,
		Enabled:               -1,
	}

	if len(req.Argv.Message) > 0 {
		err := json.Unmarshal([]byte(req.Argv.Message), &params)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "解析Json失败-"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
	}

	if len(params.AssetName) <= 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"asset_name\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if params.IsToken < 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"is_token\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if params.IsToken == 1 && len(params.ParentName) <= 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"parent_name\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if params.ConfirmationNum < 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"confirmation_num\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if params.Decimals < 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"decimals\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if params.Enabled < 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"enabled\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	assetProperty := AssetProperty{
		AssetName:             params.AssetName,
		FullName:              params.FullName,
		IsToken:               params.IsToken,
		ParentName:            params.ParentName,
		Logo:                  params.Logo,
		DepositMin:            params.DepositMin,
		WithdrawalRate:        params.WithdrawalRate,
		WithdrawalValue:       params.WithdrawalValue,
		WithdrawalReserveRate: params.WithdrawalReserveRate,
		WithdrawalAlertRate:   params.WithdrawalAlertRate,
		WithdrawalStategy:     params.WithdrawalStategy,
		ConfirmationNum:       params.ConfirmationNum,
		Decimals:              params.Decimals,
		GasFactor:             params.GasFactor,
		Debt:                  params.Debt,
		ParkAmount:            params.ParkAmount,
		Enabled:               params.Enabled,
	}
	return mysqlpool.SetAssetProperty(&assetProperty)
}
