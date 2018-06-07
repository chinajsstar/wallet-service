package chain

import (
	"bastionpay_api/api/v1"
	"bastionpay_base/data"
	"blockchain_server/service"
	. "business/def"
	"business/monitor"
	"business/mysqlpool"
	"encoding/json"
	"errors"
	"fmt"
	l4g "github.com/alecthomas/log4go"
)

func Withdrawal(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userProperty.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if userProperty.UserClass != 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "该用户不能执行该操作")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqWithdrawal{
		AssetName:   "",
		Amount:      -1,
		Address:     "",
		UserOrderID: "",
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

	if params.Amount <= 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"amount\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if len(params.Address) <= 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"address\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	userAccount, ok := mysqlpool.QueryUserAccountRow(userProperty.UserKey, assetProperty.AssetName)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "获取帐户信息失败")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if userAccount.AvailableAmount < params.Amount {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "帐户可用资金不足")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	payFee := assetProperty.WithdrawalValue + params.Amount*assetProperty.WithdrawalRate
	if userAccount.AvailableAmount < params.Amount+payFee {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "帐户可用资金不足")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if params.Amount <= 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, fmt.Sprint("提币金额要大于0"))
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	userAddress, ok := mysqlpool.QueryPayAddress(assetProperty.AssetName)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "没有设置可用的热钱包")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if userAddress.AvailableAmount < params.Amount+payFee {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "热钱包可用资金不足")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	uuID := monitor.GenerateUUID("WD")
	if len(params.UserOrderID) > 0 {
		err := mysqlpool.AddUserOrder(userProperty.UserKey, params.UserOrderID, uuID)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "不能发起重复订单交易")
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
	}

	err := mysqlpool.WithDrawalOrder(userProperty.UserKey, assetProperty.AssetName, params.Address, params.Amount,
		payFee, uuID, params.UserOrderID)
	if err != nil {
		mysqlpool.RemoveUserOrder(userProperty.UserKey, params.UserOrderID)
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "帐户可用资金不足!")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if assetProperty.IsToken > 0 {
		cmdTx, err := service.NewSendTxCmd(uuID, assetProperty.ParentName, userAddress.PrivateKey,
			params.Address, assetProperty.AssetName, userAddress.PrivateKey, params.Amount)
		if err != nil {
			mysqlpool.RemoveUserOrder(userProperty.UserKey, params.UserOrderID)
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "指令执行失败:"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		wallet.SendTx(cmdTx)
	} else {
		cmdTx, err := service.NewSendTxCmd(uuID, assetProperty.AssetName, userAddress.PrivateKey,
			params.Address, "", "", params.Amount)
		if err != nil {
			mysqlpool.RemoveUserOrder(userProperty.UserKey, params.UserOrderID)
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "指令执行失败:"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		wallet.SendTx(cmdTx)
	}

	ack := v1.AckWithdrawal{
		OrderID:     uuID,
		UserOrderID: params.UserOrderID,
	}

	pack, err := json.Marshal(ack)
	if err != nil {
		mysqlpool.RemoveUserOrder(userProperty.UserKey, params.UserOrderID)
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}
