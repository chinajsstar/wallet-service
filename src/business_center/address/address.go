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

	payFee := assetProperty.WithdrawalValue + amount*assetProperty.WithdrawalRate
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
			address, assetProperty.AssetName, userAddress.PrivateKey, amount)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "指令执行失败")
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		a.wallet.SendTx(cmdTx)
	} else {
		cmdTx, err := service.NewSendTxCmd(uuID, assetProperty.AssetName, userAddress.PrivateKey,
			address, "", "", amount)
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
		dataList.TotalLines = mysqlpool.QueryTransactionOrderCount(queryMap)
	}

	if arr, ok := mysqlpool.QueryAssetProperty(queryMap); ok {
		for _, v := range arr {
			data := v1.AckAssetsAttribute{}
			data.AssetName = v.AssetName
			data.FullName = v.FullName
			data.IsToken = v.IsToken
			data.ParentName = v.ParentName
			data.DepositMin = v.DepositMin
			data.WithdrawalRate = v.WithdrawalRate
			data.WithdrawalValue = v.WithdrawalValue
			data.ConfirmationNum = v.ConfirmationNum
			data.Decimals = v.Decimals
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

func (a *Address) GetBalance(req *data.SrvRequest, res *data.SrvResponse) error {
	_, _, userKey := req.GetUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if userProperty.UserClass != 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "此不能操作此命令")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqUserBalance{}
	if len(req.Argv.Message) > 0 {
		err := json.Unmarshal([]byte(req.Argv.Message), &params)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "解析Json失败-"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
	}

	queryMap := make(map[string]interface{})
	queryMap["user_key"] = userProperty.UserKey

	if len(params.AssetNames) > 0 {
		queryMap["asset_names"] = params.AssetNames
	}

	dataList := v1.AckUserBalanceList{}
	if arr, ok := mysqlpool.QueryUserAccount(queryMap); ok {
		for _, v := range arr {
			data := v1.AckUserBalance{}
			data.AssetName = v.AssetName
			data.AvailableAmount = v.AvailableAmount
			data.FrozenAmount = v.FrozenAmount
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

func (a *Address) HistoryTransactionOrder(req *data.SrvRequest, res *data.SrvResponse) error {
	_, _, userKey := req.GetUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqHistoryTransactionOrder{
		ID:            -1,
		OrderID:       "",
		AssetName:     "",
		TransType:     -1,
		Hash:          "",
		MaxAmount:     -1,
		MinAmount:     -1,
		MaxUpdateTime: -1,
		MinUpdateTime: -1,
		TotalLines:    -1,
		PageIndex:     -1,
		MaxDispLines:  -1,
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
	queryMap["user_key"] = userProperty.UserKey

	if params.ID > 0 {
		queryMap["id"] = params.ID
	}

	if len(params.OrderID) > 0 {
		queryMap["order_id"] = params.OrderID
	}

	if len(params.AssetName) > 0 {
		queryMap["asset_name"] = params.AssetName
	}

	if params.TransType > 0 {
		queryMap["trans_type"] = params.TransType
	}

	if len(params.Hash) > 0 {
		queryMap["hash"] = params.Hash
	}

	if params.MaxAmount > 0 {
		queryMap["max_amount"] = params.MaxAmount
	}

	if params.MinAmount > 0 {
		queryMap["min_amount"] = params.MinAmount
	}

	if params.MaxUpdateTime > 0 {
		queryMap["max_update_time"] = params.MaxUpdateTime
	}

	if params.MinUpdateTime > 0 {
		queryMap["min_update_time"] = params.MinUpdateTime
	}

	dataList := v1.AckHistoryTransactionOrderList{
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
		dataList.TotalLines = mysqlpool.QueryTransactionOrderCount(queryMap)
	}

	if arr, ok := mysqlpool.QueryTransactionOrder(queryMap); ok {
		for _, v := range arr {
			data := v1.AckHistoryTransactionOrder{}
			data.ID = v.ID
			data.AssetName = v.AssetName
			data.Address = v.Address
			data.TransType = v.TransType
			data.Status = v.Status
			data.Amount = v.Amount
			data.PayFee = v.PayFee
			data.Balance = v.Balance
			data.Hash = v.Hash
			data.OrderID = v.OrderID
			data.Time = v.Time
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

func (a *Address) HistoryTransactionMessage(req *data.SrvRequest, res *data.SrvResponse) error {
	_, _, userKey := req.GetUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqHistoryTransactionMessage{
		MaxMessageID: -1,
		MinMessageID: -1,
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
	queryMap["user_key"] = userProperty.UserKey

	if params.MaxMessageID > 0 {
		queryMap["max_msg_id"] = params.MaxMessageID
	}

	if params.MinMessageID > 0 {
		queryMap["min_msg_id"] = params.MinMessageID
	}

	dataList := v1.AckHistoryTransactionMessageList{
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
		dataList.TotalLines = mysqlpool.QueryTransactionOrderCount(queryMap)
	}

	if arr, ok := mysqlpool.QueryTransactionMessage(queryMap); ok {
		for _, v := range arr {
			data := v1.AckHistoryTransactionMessage{}
			data.MsgID = v.MsgID
			data.TransType = v.TransType
			data.Status = v.Status
			data.BlockinHeight = v.BlockinHeight
			data.AssetName = v.AssetName
			data.Address = v.Address
			data.Amount = v.Amount
			data.PayFee = v.PayFee
			data.Balance = v.Balance
			data.Hash = v.Hash
			data.OrderId = v.OrderID
			data.Time = v.Time
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

func (a *Address) QueryUserAddress(req *data.SrvRequest, res *data.SrvResponse) error {
	_, _, userKey := req.GetUserKey()
	_, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqUserAddress{
		MaxAllocationTime: -1,
		MinAllocationTime: -1,
		Address:           "",
		TotalLines:        -1,
		PageIndex:         -1,
		MaxDispLines:      -1,
	}

	err := json.Unmarshal([]byte(req.Argv.Message), &params)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, err.Error())
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	queryMap := make(map[string]interface{})
	queryMap["user_key"] = userKey

	if len(params.AssetNames) > 0 {
		queryMap["asset_names"] = params.AssetNames
	}

	if params.MaxAllocationTime > 0 {
		queryMap["max_allocation_time"] = params.MaxAllocationTime
	}

	if params.MinAllocationTime > 0 {
		queryMap["min_allocation_time"] = params.MinAllocationTime
	}

	if len(params.Address) > 0 {
		queryMap["address"] = params.Address
	}

	dataList := v1.AckUserAddressList{
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
	}

	if arr, ok := mysqlpool.QueryUserAddress(queryMap); ok {
		for _, v := range arr {
			data := v1.AckUserAddress{}
			data.AssetName = v.AssetName
			data.Address = v.Address
			data.AllocationTime = v.AllocationTime

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
