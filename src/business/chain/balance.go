package chain

import (
	"api_router/base/data"
	"bastionpay_api/apibackend/v1/backend"
	"blockchain_server/service"
	. "business/def"
	"business/mysqlpool"
	"encoding/json"
	"errors"
	l4g "github.com/alecthomas/log4go"
	"golang.org/x/net/context"
)

func SpGetChainBalance(wallet *service.ClientManager, req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if !(userProperty.UserClass == 1 || userProperty.UserClass == 2) {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "此不能操作此命令")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := backend.SpReqChainBalance{
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

	if len(params.AssetName) > 0 {
		queryMap["asset_name"] = params.AssetName
	}

	if len(params.Address) > 0 {
		queryMap["address"] = params.Address
	}

	dataList := backend.SpAckChainBalanceList{
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

	if params.TotalLines == 0 {
		dataList.TotalLines = mysqlpool.QueryUserAddressCount(queryMap)
	} else if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	}

	if arr, ok := mysqlpool.QueryUserAddress(queryMap); ok {
		for _, v := range arr {
			d := backend.SpAckChainBalance{}
			d.AssetName = v.AssetName
			d.Address = v.Address
			d.Amount = 0
			if assetProperty, ok := mysqlpool.QueryAssetPropertyByName(v.AssetName); ok {
				if assetProperty.IsToken == 0 {
					cmdBalance := service.NewQueryBalanceCmd("", assetProperty.AssetName, d.Address, "")
					balance, err := wallet.GetBalance(context.TODO(), cmdBalance, nil)
					if err == nil {
						d.Amount = balance
					}
				} else {
					cmdBalance := service.NewQueryBalanceCmd("", assetProperty.ParentName, d.Address, assetProperty.AssetName)
					balance, err := wallet.GetBalance(context.TODO(), cmdBalance, nil)
					if err == nil {
						d.Amount = balance
					}
				}
			}
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
