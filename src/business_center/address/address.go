package address

import (
	"api_router/base/data"
	"blockchain_server/service"
	"blockchain_server/types"
	. "business_center/def"
	"business_center/jsonparse"
	"business_center/mysqlpool"
	"context"
	"encoding/json"
	"errors"
	l4g "github.com/alecthomas/log4go"
	"math"
	"sync"
)

type Address struct {
	wallet          *service.ClientManager
	callback        PushMsgCallback
	rechargeChannel types.RechargeTxChannel
	cmdTxChannel    types.CmdTxChannel
	waitGroup       sync.WaitGroup
	ctx             context.Context
}

func (a *Address) Run(ctx context.Context, wallet *service.ClientManager, callback PushMsgCallback) {
	a.wallet = wallet
	a.callback = callback
	a.ctx = ctx

	a.rechargeChannel = make(types.RechargeTxChannel)
	a.cmdTxChannel = make(types.CmdTxChannel)

	a.recvRechargeTxChannel()
	a.recvCmdTxChannel()

	a.wallet.SubscribeTxRecharge(a.rechargeChannel)
	a.wallet.SubscribeTxCmdState(a.cmdTxChannel)

	//添加监控地址
	if userAddress, ok := mysqlpool.QueryUserAddress(nil); ok {
		for _, v := range userAddress {
			if assetProperty, ok := mysqlpool.QueryAssetPropertyByName(v.AssetName); ok {
				rcaCmd := service.NewRechargeAddressCmd("", assetProperty.AssetName, []string{v.Address})
				a.wallet.InsertRechargeAddress(rcaCmd)
			}
		}
	}
}

func (a *Address) Stop() {
	a.waitGroup.Wait()
}

func (a *Address) NewAddress(req *data.SrvRequestData, res *data.SrvResponseData) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"user_key\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	resMap := make(map[string]interface{})
	params := jsonparse.Parse(req.Data.Argv.Message)
	assetName, ok := params.AssetName()
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "缺少\"asset_name\"参数")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	assetProperty, ok := mysqlpool.QueryAssetPropertyByName(assetName)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"asset_name\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}
	resMap["asset_name"] = assetProperty.AssetName

	count, ok := params.Count()
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "缺少\"count\"参数")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	if count <= 0 {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"count\"要大于0")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	userAddress := a.generateAddress(&userProperty, &assetProperty, count)
	if len(userAddress) > 0 {
		data := make([]string, 0)
		for _, v := range userAddress {
			data = append(data, v.Address)
		}
		resMap["data"] = data
	}
	res.Data.Value.Message = responseJson(resMap)

	return nil
}

func (a *Address) Withdrawal(req *data.SrvRequestData, res *data.SrvResponseData) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"user_key\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	resMap := make(map[string]interface{})
	params := jsonparse.Parse(req.Data.Argv.Message)

	assetName, ok := params.AssetName()
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "缺少\"asset_name\"参数")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	amount, ok := params.Amount()
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "缺少\"amount\"参数")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	address, ok := params.Address()
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "缺少\"address\"参数")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	assetProperty, ok := mysqlpool.QueryAssetPropertyByName(assetName)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "获取币种信息失败")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	userAccount, ok := mysqlpool.QueryUserAccountRow(userProperty.UserKey, assetProperty.AssetName)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "获取帐户信息失败")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	if userAccount.AvailableAmount < amount {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "帐户可用资金不足")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	payFee := int64(assetProperty.WithdrawalValue*math.Pow10(8)) + int64(float64(amount)*assetProperty.WithdrawalRate)
	if userAccount.AvailableAmount < amount+payFee {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "帐户可用资金不足")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	userAddress, ok := mysqlpool.QueryPayAddress(assetProperty.AssetName)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorWallet, "没有设置可用的热钱包")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	if userAddress.AvailableAmount < amount+payFee {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorWallet, "热钱包资金不足(这里需要特殊处理)")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	uuID := a.generateUUID()
	resMap["order_id"] = uuID

	err := mysqlpool.WithDrawalSet(userProperty.UserKey, assetProperty.AssetName, address,
		amount, payFee, uuID)
	if err != nil {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "帐户可用资金不足!")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	pack, err := json.Marshal(resMap)
	if err != nil {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "返回数据包错误")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	if assetProperty.IsToken > 0 {
		a.wallet.SendTx(service.NewSendTxCmd(uuID, assetProperty.ParentName, userAddress.PrivateKey,
			address, &assetProperty.AssetName, uint64(amount)))
	} else {
		a.wallet.SendTx(service.NewSendTxCmd(uuID, assetProperty.AssetName, userAddress.PrivateKey,
			address, nil, uint64(amount)))
	}
	res.Data.Value.Message = string(pack)
	return nil
}

func (a *Address) SupportAssets(req *data.SrvRequestData, res *data.SrvResponseData) error {
	_, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"user_key\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	if assetProperty, ok := mysqlpool.QueryAssetProperty(nil); ok {
		data := make([]string, 0)
		for _, v := range assetProperty {
			data = append(data, v.AssetName)
		}

		pack, err := json.Marshal(data)
		if err != nil {
			res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "返回数据包错误")
			l4g.Error(res.Data.ErrMsg)
			return errors.New(res.Data.ErrMsg)
		}
		res.Data.Value.Message = string(pack)
	}
	return nil
}

func (a *Address) AssetAttribute(req *data.SrvRequestData, res *data.SrvResponseData) error {
	_, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"user_key\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	params := jsonparse.Parse(req.Data.Argv.Message)
	assetNameArray, _ := params.AssetNameArray()

	assetPropertyMap := make(map[string]map[string]interface{})
	if assetProperty, ok := mysqlpool.QueryAssetProperty(nil); ok {
		for _, v := range assetProperty {
			maps := make(map[string]interface{}, 0)
			maps["asset_name"] = v.AssetName
			maps["full_name"] = v.FullName
			maps["is_token"] = v.IsToken
			maps["parent_name"] = v.ParentName
			maps["deposit_min"] = v.DepositMin
			maps["withdrawal_rate"] = v.WithdrawalRate
			maps["withdrawal_value"] = v.WithdrawalValue
			maps["confirmation_num"] = v.ConfirmationNum
			maps["decimal"] = v.Decimal
			assetPropertyMap[v.AssetName] = maps
		}
	}

	var data []map[string]interface{}
	if len(assetNameArray) <= 0 {
		for _, value := range assetPropertyMap {
			data = append(data, value)
		}
	} else {
		for _, value := range assetNameArray {
			if value, ok := assetPropertyMap[value]; ok {
				data = append(data, value)
			}
		}
	}

	pack, err := json.Marshal(data)
	if err != nil {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "返回数据包错误")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}
	res.Data.Value.Message = string(pack)
	return nil
}

func (a *Address) GetBalance(req *data.SrvRequestData, res *data.SrvResponseData) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"user_key\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	queryMap := make(map[string]interface{})
	params := jsonparse.Parse(req.Data.Argv.Message)

	if userProperty.UserClass == 1 {
		if value, ok := params.UserKey(); ok {
			queryMap["user_key"] = value
		}

	} else {
		queryMap["user_key"] = req.Data.Argv.UserKey
	}

	userAccountMap := make(map[string]map[string]interface{})
	if userAccount, ok := mysqlpool.QueryUserAccount(queryMap); ok {
		for _, v := range userAccount {
			maps := make(map[string]interface{}, 0)
			maps["asset_name"] = v.AssetName
			maps["available_amount"] = float64(v.AvailableAmount) * math.Pow10(-8)
			maps["frozen_amount"] = float64(v.FrozenAmount) * math.Pow10(-8)
			userAccountMap[v.AssetName] = maps
		}
	}

	var data []map[string]interface{}
	if assetName, ok := params.AssetNameArray(); ok {
		for _, v := range assetName {
			if value, ok := userAccountMap[v]; ok {
				data = append(data, value)
			}
		}
	} else {
		for _, v := range userAccountMap {
			data = append(data, v)
		}
	}

	pack, err := json.Marshal(data)
	if err != nil {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "返回数据包错误")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}
	res.Data.Value.Message = string(pack)
	return nil
}

func (a *Address) HistoryTransactionOrder(req *data.SrvRequestData, res *data.SrvResponseData) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"user_key\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	queryMap := make(map[string]interface{})
	params := jsonparse.Parse(req.Data.Argv.Message)

	if userProperty.UserClass == 1 {
		if value, ok := params.UserKey(); ok {
			queryMap["user_key"] = value
		}

	} else {
		queryMap["user_key"] = req.Data.Argv.UserKey
	}

	if value, ok := params.AssetName(); ok {
		queryMap["asset_name"] = value
	}

	if value, ok := params.TransType(); ok {
		queryMap["trans_type"] = value
	}

	if value, ok := params.Status(); ok {
		queryMap["status"] = value
	}

	if value, ok := params.MaxUpdateTime(); ok {
		queryMap["max_update_time"] = value
	}

	if value, ok := params.MinUpdateTime(); ok {
		queryMap["min_update_time"] = value
	}

	data, _ := mysqlpool.QueryTransactionOrder(queryMap)
	pack, err := json.Marshal(data)
	if err != nil {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "返回数据包错误")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}
	res.Data.Value.Message = string(pack)
	return nil
}

func (a *Address) HistoryTransactionMessage(req *data.SrvRequestData, res *data.SrvResponseData) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"user_key\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	queryMap := make(map[string]interface{})
	params := jsonparse.Parse(req.Data.Argv.Message)

	if userProperty.UserClass == 1 {
		if value, ok := params.UserKey(); ok {
			queryMap["user_key"] = value
		}

	} else {
		queryMap["user_key"] = req.Data.Argv.UserKey
	}

	if value, ok := params.AssetName(); ok {
		queryMap["max_msg_id"] = value
	}

	if value, ok := params.TransType(); ok {
		queryMap["min_msg_id"] = value
	}

	data, _ := mysqlpool.QueryTransactionMessage(queryMap)
	pack, err := json.Marshal(data)
	if err != nil {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "返回数据包错误")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}
	res.Data.Value.Message = string(pack)
	return nil
}

func (a *Address) QueryAssetProperty(req *data.SrvRequestData, res *data.SrvResponseData) error {
	resMap := responsePagination(nil, mysqlpool.QueryAssetPropertyCount(nil))
	assetProperty, _ := mysqlpool.QueryAssetProperty(nil)
	resMap["data"] = assetProperty

	res.Data.Value.Message = responseJson(resMap)
	res.Data.Err = 0
	res.Data.ErrMsg = ""
	return nil
}

func (a *Address) QueryUserProperty(req *data.SrvRequestData, res *data.SrvResponseData) error {
	resMap := responsePagination(nil, mysqlpool.QueryUserPropertyCount(nil))
	userProperty, _ := mysqlpool.QueryUserProperty(nil)
	resMap["data"] = userProperty

	res.Data.Value.Message = responseJson(resMap)
	res.Data.Err = 0
	res.Data.ErrMsg = ""
	return nil
}

func (a *Address) QueryUserAccount(req *data.SrvRequestData, res *data.SrvResponseData) error {
	resMap := responsePagination(nil, mysqlpool.QueryUserAccountCount(nil))
	userAccount, _ := mysqlpool.QueryUserAccount(nil)
	resMap["data"] = userAccount

	res.Data.Value.Message = responseJson(resMap)
	res.Data.Err = 0
	res.Data.ErrMsg = ""
	return nil
}

func (a *Address) QueryUserAddress(req *data.SrvRequestData, res *data.SrvResponseData) error {
	resMap := responsePagination(nil, mysqlpool.QueryUserAddressCount(nil))
	userAddress, _ := mysqlpool.QueryUserAddress(nil)
	resMap["data"] = userAddress

	res.Data.Value.Message = responseJson(resMap)
	res.Data.Err = 0
	res.Data.ErrMsg = ""
	return nil
}

func (a *Address) SetPayAddress(req *data.SrvRequestData, res *data.SrvResponseData) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"user_key\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	if userProperty.UserClass != 1 {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParam, "不能操作此命令")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	params := jsonparse.Parse(req.Data.Argv.Message)
	assetName, ok := params.AssetName()
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "缺少\"asset_name\"参数")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	address, ok := params.Address()
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "缺少\"address\"参数")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	err := mysqlpool.SetPayAddress(assetName, address)
	if err != nil {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, err.Error())
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}
	return nil
}
