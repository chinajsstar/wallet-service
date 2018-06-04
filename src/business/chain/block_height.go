package chain

import (
	"bastionpay_base/data"
	"bastionpay_api/api/v1"
	. "business/def"
	"business/mysqlpool"
	"encoding/json"
	"errors"
	l4g "github.com/alecthomas/log4go"
	"time"
)

func BlockHeight(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userProperty.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqBlockHeight{}
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

	dataList := v1.AckBlockHeightList{}
	if assetProperty, ok := mysqlpool.QueryAssetProperty(queryMap); ok {
		for _, v := range assetProperty {
			data := v1.AckBlockHeight{}
			data.AssetName = v.AssetName
			data.BlockHeight = 0
			if v.IsToken > 0 {
				data.BlockHeight = int64(wallet.BlockHeight(v.ParentName))
			} else {
				data.BlockHeight = int64(wallet.BlockHeight(v.AssetName))
			}
			data.UpdateTime = time.Now().UTC().Unix()
			dataList.Data = append(dataList.Data, data)
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
