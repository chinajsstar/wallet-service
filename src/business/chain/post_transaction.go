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
	"strconv"
)

func SpPostTransaction(req *data.SrvRequest, res *data.SrvResponse) error {
	params := backend.SpReqPostTransaction{
		AssetName: "",
		From:      "",
		To:        "",
		Amount:    -1,
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

	assetProperty, ok := mysqlpool.QueryAssetPropertyByName(params.AssetName)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "参数:\"asset_name\"无效")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if len(params.From) <= 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"From\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	userAddress, ok := mysqlpool.QueryUserAddressByNameAddress(assetProperty.AssetName, params.From)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "\"From\"参数无效")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if params.Amount > userAddress.AvailableAmount {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "\"From\"地址金额不足:"+strconv.FormatFloat(userAddress.AvailableAmount, 'f', -1, 64))
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if assetProperty.IsToken > 0 {
		hotAddress, ok := mysqlpool.QueryPayAddress(assetProperty.AssetName)
		if !ok {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "没有设置可用的热钱包")
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}

		if params.To == hotAddress.Address {
			cmdTx, err := service.NewSendTxCmd("", assetProperty.ParentName, hotAddress.PrivateKey,
				params.To, assetProperty.AssetName, userAddress.PrivateKey, params.Amount)
			if err != nil {
				res.Err, res.ErrMsg = CheckError(ErrorFailed, "指令执行失败:"+err.Error())
				l4g.Error(res.ErrMsg)
				return errors.New(res.ErrMsg)
			}
			wallet.SendTx(cmdTx)
		} else {
			cmdTx, err := service.NewSendTxCmd("", assetProperty.ParentName, userAddress.PrivateKey,
				params.To, assetProperty.AssetName, userAddress.PrivateKey, params.Amount)
			if err != nil {
				res.Err, res.ErrMsg = CheckError(ErrorFailed, "指令执行失败:"+err.Error())
				l4g.Error(res.ErrMsg)
				return errors.New(res.ErrMsg)
			}
			wallet.SendTx(cmdTx)
		}
	} else {
		cmdTx, err := service.NewSendTxCmd("", assetProperty.AssetName, userAddress.PrivateKey,
			params.To, "", "", params.Amount)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "指令执行失败:"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		wallet.SendTx(cmdTx)
	}
	return nil
}
