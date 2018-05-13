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
				assetName := assetProperty.AssetName
				if assetProperty.IsToken > 0 {
					assetName = assetProperty.ParentName
				}
				rcaCmd := service.NewRechargeAddressCmd("", assetName, []string{v.Address})
				a.wallet.InsertRechargeAddress(rcaCmd)
			}
		}
	}
}

func (a *Address) Stop() {
	a.waitGroup.Wait()
}

func (a *Address) NewAddress(req *data.SrvRequest, res *data.SrvResponse) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Argv.UserKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+req.Argv.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	resMap := make(map[string]interface{})
	params, err := jsonparse.Parse(req.Argv.Message)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	assetName, ok := params.AssetName()
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"asset_name\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	assetProperty, ok := mysqlpool.QueryAssetPropertyByName(assetName)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "参数:\"asset_name\"无效")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	resMap["asset_name"] = assetProperty.AssetName

	count, ok := params.Count()
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"count\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if count <= 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "参数:\"count\"要大于0")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	userAddress := a.generateAddress(&userProperty, &assetProperty, count)
	if len(userAddress) > 0 {
		data := make([]string, 0)
		for _, v := range userAddress {
			data = append(data, v.Address)
		}
		resMap["data"] = data
	}
	res.Value.Message = responseJson(resMap)

	return nil
}

func (a *Address) Withdrawal(req *data.SrvRequest, res *data.SrvResponse) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Argv.UserKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+req.Argv.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	resMap := make(map[string]interface{})
	params, err := jsonparse.Parse(req.Argv.Message)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "解析Json失败-"+err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	assetName, ok := params.AssetName()
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"asset_name\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	amount, ok := params.Amount()
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"amount\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	address, ok := params.Address()
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"address\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	assetProperty, ok := mysqlpool.QueryAssetPropertyByName(assetName)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "获取币种信息失败")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	userAccount, ok := mysqlpool.QueryUserAccountRow(userProperty.UserKey, assetProperty.AssetName)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "获取帐户信息失败")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if userAccount.AvailableAmount < amount {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "帐户可用资金不足")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	payFee := int64(assetProperty.WithdrawalValue*math.Pow10(8)) + int64(float64(amount)*assetProperty.WithdrawalRate)
	if userAccount.AvailableAmount < amount+payFee {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "帐户可用资金不足")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	userAddress, ok := mysqlpool.QueryPayAddress([]string{assetProperty.AssetName})
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "没有设置可用的热钱包")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if userAddress.AvailableAmount < amount+payFee {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "热钱包可用资金不足(这里需要特殊处理)")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	uuID := a.generateUUID()
	resMap["order_id"] = uuID

	err = mysqlpool.WithDrawalSet(userProperty.UserKey, assetProperty.AssetName, address, amount, payFee, uuID)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "帐户可用资金不足!")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	pack, err := json.Marshal(resMap)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if assetProperty.IsToken > 0 {
		a.wallet.SendTx(service.NewSendTxCmd(uuID, assetProperty.ParentName, userAddress.PrivateKey,
			address, &assetProperty.AssetName, uint64(amount)))
	} else {
		a.wallet.SendTx(service.NewSendTxCmd(uuID, assetProperty.AssetName, userAddress.PrivateKey,
			address, nil, uint64(amount)))
	}
	res.Value.Message = string(pack)
	return nil
}

func (a *Address) SupportAssets(req *data.SrvRequest, res *data.SrvResponse) error {
	_, ok := mysqlpool.QueryUserPropertyByKey(req.Argv.UserKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+req.Argv.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if assetProperty, ok := mysqlpool.QueryAssetProperty(nil); ok {
		data := make([]string, 0)
		for _, v := range assetProperty {
			data = append(data, v.AssetName)
		}

		pack, err := json.Marshal(data)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		res.Value.Message = string(pack)
	}
	return nil
}

func (a *Address) AssetAttribute(req *data.SrvRequest, res *data.SrvResponse) error {
	_, ok := mysqlpool.QueryUserPropertyByKey(req.Argv.UserKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+req.Argv.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params, err := jsonparse.Parse(req.Argv.Message)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

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
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func (a *Address) GetBalance(req *data.SrvRequest, res *data.SrvResponse) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Argv.UserKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+req.Argv.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	queryMap := make(map[string]interface{})
	params, err := jsonparse.Parse(req.Argv.Message)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if userProperty.UserClass == 1 {
		if value, ok := params.UserKey(); ok {
			queryMap["user_key"] = value
		}

	} else {
		queryMap["user_key"] = req.Argv.UserKey
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
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func (a *Address) HistoryTransactionOrder(req *data.SrvRequest, res *data.SrvResponse) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Argv.UserKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+req.Argv.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	queryMap := make(map[string]interface{})
	params, err := jsonparse.Parse(req.Argv.Message)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if userProperty.UserClass == 1 {
		if value, ok := params.UserKey(); ok {
			queryMap["user_key"] = value
		}

	} else {
		queryMap["user_key"] = req.Argv.UserKey
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
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func (a *Address) HistoryTransactionMessage(req *data.SrvRequest, res *data.SrvResponse) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Argv.UserKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+req.Argv.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	queryMap := make(map[string]interface{})
	params, err := jsonparse.Parse(req.Argv.Message)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if userProperty.UserClass == 1 {
		if value, ok := params.UserKey(); ok {
			queryMap["user_key"] = value
		}

	} else {
		queryMap["user_key"] = req.Argv.UserKey
	}

	if value, ok := params.MaxMessageID(); ok {
		queryMap["max_msg_id"] = value
	}

	if value, ok := params.MinMessageID(); ok {
		queryMap["min_msg_id"] = value
	}

	data, _ := mysqlpool.QueryTransactionMessage(queryMap)
	pack, err := json.Marshal(data)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func (a *Address) QueryUserAddress(req *data.SrvRequest, res *data.SrvResponse) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Argv.UserKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+req.Argv.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	queryMap := make(map[string]interface{})
	params, err := jsonparse.Parse(req.Argv.Message)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if userProperty.UserClass == 1 {
		if value, ok := params.UserKey(); ok {
			queryMap["user_key"] = value
		}

	} else {
		queryMap["user_key"] = req.Argv.UserKey
	}

	if value, ok := params.AssetName(); ok {
		queryMap["asset_name"] = value
	}

	resMap := responsePagination(queryMap, mysqlpool.QueryUserAddressCount(queryMap))
	userAddress, _ := mysqlpool.QueryUserAddress(queryMap)
	resMap["data"] = userAddress

	res.Value.Message = responseJson(resMap)
	res.Err = 0
	res.ErrMsg = ""
	return nil
}

func (a *Address) SetPayAddress(req *data.SrvRequest, res *data.SrvResponse) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Argv.UserKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+req.Argv.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if userProperty.UserClass != 1 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "此不能操作此命令")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params, err := jsonparse.Parse(req.Argv.Message)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	assetName, ok := params.AssetName()
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"asset_name\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	address, ok := params.Address()
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "缺少\"address\"参数")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	err = mysqlpool.SetPayAddress(assetName, address)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	return nil
}

func (a *Address) QueryPayAddress(req *data.SrvRequest, res *data.SrvResponse) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Argv.UserKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+req.Argv.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if userProperty.UserClass != 1 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "此不能操作此命令")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params, err := jsonparse.Parse(req.Argv.Message)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	assetName, _ := params.AssetNameArray()
	userAddress, _ := mysqlpool.QueryPayAddress(assetName)
	res.Value.Message = responseJson(userAddress)
	res.Err = 0
	res.ErrMsg = ""

	return nil
}
