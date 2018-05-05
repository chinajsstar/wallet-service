package address

import (
	"api_router/base/data"
	"blockchain_server/service"
	"blockchain_server/types"
	. "business_center/def"
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
	if userAddress, ok := mysqlpool.QueryUserAddress(""); ok {
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

	paramsMapping := unpackJson(req.Data.Argv.Message)
	resMap := make(map[string]interface{})
	resMap["asset_name"] = paramsMapping.AssetName

	assetProperty, ok := mysqlpool.QueryAssetPropertyByName(paramsMapping.AssetName)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"asset_name\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	if paramsMapping.Count <= 0 {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"count\"要大于0")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	userAddress := a.generateAddress(&userProperty, &assetProperty, paramsMapping.Count)
	if len(userAddress) > 0 {
		data := make([]string, 0)
		for _, v := range userAddress {
			data = append(data, v.Address)
		}
		resMap["data"] = data
	}
	res.Data.Value.Message = packJson(resMap)

	return nil
}

func (a *Address) Withdrawal(req *data.SrvRequestData, res *data.SrvResponseData) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"user_key\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	paramsMapping := unpackJson(req.Data.Argv.Message)
	resMap := make(map[string]interface{})

	assetProperty, ok := mysqlpool.QueryAssetPropertyByName(paramsMapping.AssetName)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "获取用户帐户信息失败")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	userAccount, ok := mysqlpool.QueryUserAccountRow(userProperty.UserKey, assetProperty.AssetName)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "获取用户帐户信息失败")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	if userAccount.AvailableAmount < paramsMapping.Amount {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "用户帐户可用资金不足")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	payFee := int64(assetProperty.WithdrawalValue*math.Pow10(8)) + int64(float64(paramsMapping.Amount)*assetProperty.WithdrawalRate)
	if userAccount.AvailableAmount < paramsMapping.Amount+payFee {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "用户帐户可用资金不足")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	userAddress, ok := mysqlpool.QueryPayAddress(assetProperty.AssetName)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorWallet, "没有设置可用的热钱包")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	if userAddress.AvailableAmount < paramsMapping.Amount+payFee {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorWallet, "热钱包资金不足，这里需要特殊处理")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	uuID := a.generateUUID()
	resMap["order_id"] = uuID

	err := mysqlpool.WithDrawalSet(userProperty.UserKey, assetProperty.AssetName, paramsMapping.Address,
		paramsMapping.Amount, payFee, uuID)
	if err != nil {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "用户帐户可用资金不足!")
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
			paramsMapping.Address, &assetProperty.AssetName, uint64(paramsMapping.Amount)))
	} else {
		a.wallet.SendTx(service.NewSendTxCmd(uuID, assetProperty.AssetName, userAddress.PrivateKey,
			paramsMapping.Address, nil, uint64(paramsMapping.Amount)))
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

	if assetProperty, ok := mysqlpool.QueryAssetPropertyByJson(""); ok {
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

func (a *Address) AssetAttributie(req *data.SrvRequestData, res *data.SrvResponseData) error {
	_, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"user_key\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	paramsMapping := unpackJson(req.Data.Argv.Message)
	assetPropertyMap := make(map[string]map[string]interface{})

	if assetProperty, ok := mysqlpool.QueryAssetPropertyByJson(""); ok {
		for _, v := range assetProperty {
			maps := make(map[string]interface{}, 0)
			maps["asset_name"] = v.AssetName
			maps["full_name"] = v.FullName
			maps["is_token"] = v.IsToken
			maps["parent_name"] = v.ParentName
			maps["deposit_min"] = v.DepositMin
			maps["withrawal_rate"] = v.WithdrawalRate
			maps["withrawal_value"] = v.WithdrawalValue
			maps["confirmation_num"] = v.ConfirmationNum
			maps["decaimal"] = v.Decaimal
			assetPropertyMap[v.AssetName] = maps
		}
	}

	var data []map[string]interface{}
	if len(paramsMapping.Params) <= 0 {
		for _, v := range assetPropertyMap {
			data = append(data, v)
		}
	} else {
		for _, v := range paramsMapping.Params {
			if value, ok := assetPropertyMap[v]; ok {
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

	paramsMapping := unpackJson(req.Data.Argv.Message)
	userAccountMap := make(map[string]map[string]interface{})

	if userAccount, ok := mysqlpool.QueryUserAccount(userProperty.UserKey, ""); ok {
		for _, v := range userAccount {
			maps := make(map[string]interface{}, 0)
			maps["asset_name"] = v.AssetName
			maps["available_amount"] = float64(v.AvailableAmount) * math.Pow10(-8)
			maps["frozen_amount"] = float64(v.FrozenAmount) * math.Pow10(-8)
			userAccountMap[v.AssetName] = maps
		}
	}

	var data []map[string]interface{}
	if len(paramsMapping.Params) <= 0 {
		for _, v := range userAccountMap {
			data = append(data, v)
		}
	} else {
		for _, v := range paramsMapping.Params {
			if value, ok := userAccountMap[v]; ok {
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

func (a *Address) HistoryTransactionOrder(req *data.SrvRequestData, res *data.SrvResponseData) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"user_key\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	jsonMap := make(map[string]interface{})
	if len(req.Data.Argv.Message) > 0 {
		err := json.Unmarshal([]byte(req.Data.Argv.Message), &jsonMap)
		if err != nil {
			res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "解析Json失败")
			l4g.Error(res.Data.ErrMsg)
			return errors.New(res.Data.ErrMsg)
		}
	}

	paramsMap := make(map[string]interface{})
	paramsMap["user_key"] = userProperty.UserKey

	if value, ok := jsonMap["asset_name"]; ok {
		paramsMap["asset_name"] = value
	}

	if value, ok := jsonMap["trans_type"]; ok {
		paramsMap["trans_type"] = value
	}

	if value, ok := jsonMap["status"]; ok {
		paramsMap["status"] = value
	}

	if value, ok := jsonMap["max_update_time"]; ok {
		paramsMap["max_update_time"] = value
	}

	if value, ok := jsonMap["min_update_time"]; ok {
		paramsMap["min_update_time"] = value
	}

	condi, err := json.Marshal(paramsMap)
	if err != nil {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "Json序列化失败")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	data, ok := mysqlpool.QueryTransactionOrderByJson(string(condi))
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "Json序列化失败")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
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

func (a *Address) HistoryTransactionMessage(req *data.SrvRequestData, res *data.SrvResponseData) error {
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"user_key\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	jsonMap := make(map[string]interface{})
	if len(req.Data.Argv.Message) > 0 {
		err := json.Unmarshal([]byte(req.Data.Argv.Message), &jsonMap)
		if err != nil {
			res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "解析Json失败")
			l4g.Error(res.Data.ErrMsg)
			return errors.New(res.Data.ErrMsg)
		}
	}

	paramsMap := make(map[string]interface{})
	paramsMap["user_key"] = userProperty.UserKey

	if value, ok := jsonMap["max_msg_id"]; ok {
		paramsMap["max_msg_id"] = value
	}

	if value, ok := jsonMap["min_msg_id"]; ok {
		paramsMap["min_msg_id"] = value
	}

	condi, err := json.Marshal(paramsMap)
	if err != nil {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "Json序列化失败")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	data, ok := mysqlpool.QueryTransactionMessageByJson(string(condi))
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "Json序列化失败")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
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

func (a *Address) QueryAssetProperty(req *data.SrvRequestData, res *data.SrvResponseData) error {
	query := req.Data.Argv.Message
	resMap := responsePagination(query, mysqlpool.QueryAssetPropertyCountByJson(query))
	assetProperty, _ := mysqlpool.QueryAssetPropertyByJson(query)
	resMap["data"] = assetProperty

	res.Data.Value.Message = packJson(resMap)
	res.Data.Err = 0
	res.Data.ErrMsg = ""
	return nil
}

func (a *Address) QueryUserProperty(req *data.SrvRequestData, res *data.SrvResponseData) error {
	query := req.Data.Argv.Message
	resMap := responsePagination(query, mysqlpool.QueryUserPropertyCountByJson(query))
	userProperty, _ := mysqlpool.QueryUserPropertyByJson(query)
	resMap["data"] = userProperty

	res.Data.Value.Message = packJson(resMap)
	res.Data.Err = 0
	res.Data.ErrMsg = ""
	return nil
}

func (a *Address) QueryUserAccount(req *data.SrvRequestData, res *data.SrvResponseData) error {
	query := req.Data.Argv.Message
	resMap := responsePagination(query, mysqlpool.QueryUserAccountCountByJson(query))
	userAccount, _ := mysqlpool.QueryUserAccountByJson(query)
	resMap["data"] = userAccount

	res.Data.Value.Message = packJson(resMap)
	res.Data.Err = 0
	res.Data.ErrMsg = ""
	return nil
}

func (a *Address) QueryUserAddress(req *data.SrvRequestData, res *data.SrvResponseData) error {
	query := req.Data.Argv.Message
	resMap := responsePagination(query, mysqlpool.QueryUserAddressCount(query))
	userAddress, _ := mysqlpool.QueryUserAddress(query)
	resMap["data"] = userAddress

	res.Data.Value.Message = packJson(resMap)
	res.Data.Err = 0
	res.Data.ErrMsg = ""
	return nil
}
