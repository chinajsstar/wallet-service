package address

import (
	"api_router/base/data"
	"bastionpay_api/api/v1"
	"blockchain_server/service"
	"blockchain_server/types"
	. "business_center/def"
	"business_center/jsonparse"
	"business_center/mysqlpool"
	"business_center/transaction"
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

	if userProperty.UserClass != 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "该用户不能执行该操作")
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

	uuID := transaction.GenerateUUID("WD")
	userOrderID, ok := params.UserOrderID()
	if ok {
		if len(userOrderID) > 0 {
			err := mysqlpool.AddUserOrder(userProperty.UserKey, userOrderID, uuID)
			if err != nil {
				res.Err, res.ErrMsg = CheckError(ErrorFailed, "不能发起重复订单交易")
				l4g.Error(res.ErrMsg)
				return errors.New(res.ErrMsg)
			}
		}
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

	if amount <= 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "提币金额要大于0")
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

	if amount <= 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "提币金额必需大于0")
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

	err = mysqlpool.WithDrawalOrder(userProperty.UserKey, assetProperty.AssetName, address, amount, payFee, uuID, userOrderID)
	if err != nil {
		mysqlpool.RemoveUserOrder(userProperty.UserKey, userOrderID)
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "帐户可用资金不足!")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if assetProperty.IsToken > 0 {
		cmdTx, err := service.NewSendTxCmd(uuID, assetProperty.ParentName, userAddress.PrivateKey,
			address, assetProperty.AssetName, userAddress.PrivateKey, float64(amount)*math.Pow10(-8))
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "指令执行失败")
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		a.wallet.SendTx(cmdTx)
	} else {
		cmdTx, err := service.NewSendTxCmd(uuID, assetProperty.AssetName, userAddress.PrivateKey,
			address, "", "", toChainValue(amount))
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "指令执行失败")
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		a.wallet.SendTx(cmdTx)
	}

	resMap["order_id"] = uuID
	resMap["user_order_id"] = userOrderID

	pack, err := json.Marshal(resMap)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func (a *Address) SupportAssets(req *data.SrvRequest, res *data.SrvResponse) error {
	_, _, userKey := req.GetUserKey()
	if _, ok := mysqlpool.QueryUserPropertyByKey(userKey); !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
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

func (a *Address) AssetAttribute(req *data.SrvRequest, res *data.SrvResponse) error {
	_, _, userKey := req.GetUserKey()
	if _, ok := mysqlpool.QueryUserPropertyByKey(userKey); !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
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

	assetNameMap := make(map[string]interface{})
	for _, value := range params.AssetNames {
		if value != "" {
			assetNameMap[value] = ""
		}
	}

	assetsAttributeList := v1.AckAssetsAttributeList{}
	if assetProperty, ok := mysqlpool.QueryAssetProperty(nil); ok {
		for _, v := range assetProperty {
			if len(assetNameMap) > 0 {
				if _, ok := assetNameMap[v.AssetName]; !ok {
					continue
				}
			}
			if params.IsToken > 0 {
				if params.IsToken != v.IsToken {
					continue
				}
			}
			assetAttribute := v1.AckAssetsAttribute{}
			assetAttribute.AssetName = v.AssetName
			assetAttribute.FullName = v.FullName
			assetAttribute.IsToken = v.IsToken
			assetAttribute.ParentName = v.ParentName
			assetAttribute.DepositMin = v.DepositMin
			assetAttribute.WithdrawalRate = v.WithdrawalRate
			assetAttribute.WithdrawalValue = v.WithdrawalValue
			assetAttribute.ConfirmationNum = v.ConfirmationNum
			assetAttribute.Decimals = v.Decimals
			assetsAttributeList.Data = append(assetsAttributeList.Data, assetAttribute)
		}
	}

	pack, err := json.Marshal(assetsAttributeList)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func (a *Address) GetBalance(req *data.SrvRequest, res *data.SrvResponse) error {
	// 获取userKey和是否sub
	_, _, realUseKey := req.GetUserKey()
	isSubUserKey := req.IsSubUserKey()

	_, ok := mysqlpool.QueryUserPropertyByKey(realUseKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+realUseKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	var assetNameArray []string
	if isSubUserKey {
		assets := v1.ReqUserBalance{}
		err := json.Unmarshal([]byte(req.Argv.Message), &assets)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		assetNameArray = assets.Assets
	} else {
		params, err := jsonparse.Parse(req.Argv.Message)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		assetNameArray, _ = params.AssetNameArray()
	}

	// 放入map
	assetNameMap := make(map[string]interface{})
	for _, value := range assetNameArray {
		if value != "" {
			assetNameMap[value] = ""
		}
	}

	queryMap := make(map[string]interface{})
	queryMap["user_key"] = realUseKey

	assetsBalanceList := v1.AckUserBalanceList{}
	if userAccount, ok := mysqlpool.QueryUserAccount(queryMap); ok {
		for _, v := range userAccount {
			// 是否选中的
			if len(assetNameMap) > 0 {
				if _, ok := assetNameMap[v.AssetName]; !ok {
					continue
				}
			}

			assetBalance := v1.AckUserBalance{}
			assetBalance.AssetName = v.AssetName
			assetBalance.AvailableAmount = float64(v.AvailableAmount) * math.Pow10(-8)
			assetBalance.FrozenAmount = float64(v.FrozenAmount) * math.Pow10(-8)

			assetsBalanceList.Data = append(assetsBalanceList.Data, assetBalance)
		}
	}

	// TODO：输出分叉，以后xuliang处理
	var err error
	var pack []byte
	if isSubUserKey {
		pack, err = json.Marshal(assetsBalanceList)
	} else {
		pack, err = json.Marshal(assetsBalanceList.Data)
	}

	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)

	return nil
}

func (a *Address) HistoryTransactionOrder(req *data.SrvRequest, res *data.SrvResponse) error {
	// 获取userKey和是否sub
	_, _, realUseKey := req.GetUserKey()
	isSubUserKey := req.IsSubUserKey()

	_, ok := mysqlpool.QueryUserPropertyByKey(realUseKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+realUseKey)
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

	queryMap["user_key"] = realUseKey

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

	var pack []byte
	if isSubUserKey {
		hisTxOrderList := v1.AckHistoryTransactionOrderList{}

		for _, v := range data {
			hisTxOrder := v1.AckHistoryTransactionOrder{}

			hisTxOrder.AssetName = v.AssetName
			hisTxOrder.TransType = v.TransType
			hisTxOrder.Status = v.Status
			hisTxOrder.Amount = v.Amount
			hisTxOrder.PayFee = v.PayFee
			hisTxOrder.Hash = v.Hash
			hisTxOrder.OrderID = v.OrderID
			hisTxOrder.Time = v.Time

			hisTxOrderList.Data = append(hisTxOrderList.Data, hisTxOrder)
		}

		pack, err = json.Marshal(hisTxOrderList)
	} else {
		pack, err = json.Marshal(data)
	}

	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func (a *Address) HistoryTransactionMessage(req *data.SrvRequest, res *data.SrvResponse) error {
	// 获取userKey和是否sub
	_, _, realUseKey := req.GetUserKey()
	isSubUserKey := req.IsSubUserKey()

	_, ok := mysqlpool.QueryUserPropertyByKey(realUseKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+realUseKey)
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

	queryMap["user_key"] = realUseKey

	if value, ok := params.MaxMessageID(); ok {
		queryMap["max_msg_id"] = value
	}

	if value, ok := params.MinMessageID(); ok {
		queryMap["min_msg_id"] = value
	}

	data, _ := mysqlpool.QueryTransactionMessage(queryMap)

	var pack []byte
	if isSubUserKey {
		hisTxMsgList := v1.AckHistoryTransactionMessageList{}

		for _, v := range data {
			hisTxMsg := v1.AckHistoryTransactionMessage{}

			hisTxMsg.MsgId = v.MsgID
			hisTxMsg.TransType = v.TransType
			hisTxMsg.Status = v.Status
			hisTxMsg.BlockinHeight = v.BlockinHeight
			hisTxMsg.AssetName = v.AssetName
			hisTxMsg.Address = v.Address
			hisTxMsg.Amount = v.Amount
			hisTxMsg.PayFee = v.PayFee
			hisTxMsg.Hash = v.Hash
			hisTxMsg.OrderId = v.OrderID
			hisTxMsg.Time = v.Time

			hisTxMsgList.Data = append(hisTxMsgList.Data, hisTxMsg)
		}

		pack, err = json.Marshal(hisTxMsgList)
	} else {
		pack, err = json.Marshal(data)
	}

	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func (a *Address) QueryUserAddress(req *data.SrvRequest, res *data.SrvResponse) error {
	// 获取userKey和是否sub
	_, _, realUseKey := req.GetUserKey()
	//isSubUserKey := req.IsSubUserKey()

	_, ok := mysqlpool.QueryUserPropertyByKey(realUseKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+realUseKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	reqUserAddrss := v1.ReqUserAddress{}
	err := json.Unmarshal([]byte(req.Argv.Message), &reqUserAddrss)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	queryMap := make(map[string]interface{})
	queryMap["user_key"] = realUseKey
	if reqUserAddrss.AssetName != "" {
		queryMap["asset_name"] = reqUserAddrss.AssetName
	}

	resMap := responsePagination(queryMap, mysqlpool.QueryUserAddressCount(queryMap))
	userAddress, _ := mysqlpool.QueryUserAddress(queryMap)

	userAddressList := v1.AckUserAddressList{}
	for _, v := range userAddress {
		ua := v1.AckUserAddress{}
		ua.AssetName = v.AssetName
		ua.Address = v.Address
		ua.AllocationTime = v.AllocationTime

		userAddressList.Data = append(userAddressList.Data, ua)
	}
	resMap["data"] = userAddressList.Data

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
