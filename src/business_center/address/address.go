package address

import (
	"api_router/base/data"
	"bastionpay_api/api/v1"
	"bastionpay_api/apibackend/v1/backend"
	"blockchain_server/service"
	"blockchain_server/types"
	. "business_center/def"
	"business_center/jsonparse"
	"business_center/mysqlpool"
	"business_center/transaction"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"strconv"
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
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userProperty.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqNewAddress{
		AssetName: "",
		Count:     -1,
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

	if params.Count <= 0 {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "参数:\"count\"要大于0")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	dataList := v1.AckNewAddressList{
		AssetName: assetProperty.AssetName,
	}

	userAddress := a.generateAddress(&userProperty, &assetProperty, params.Count)
	if len(userAddress) > 0 {
		for _, v := range userAddress {
			dataList.Data = append(dataList.Data, v.Address)
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

func (a *Address) Withdrawal(req *data.SrvRequest, res *data.SrvResponse) error {
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

	uuID := transaction.GenerateUUID("WD")
	if len(params.UserOrderID) > 0 {
		err := mysqlpool.AddUserOrder(userProperty.UserKey, params.UserOrderID, uuID)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "不能发起重复订单交易")
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

	if params.Amount <= assetProperty.DepositMin {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, fmt.Sprint("提币金额要大于", assetProperty.DepositMin))
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
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "热钱包可用资金不足(这里需要特殊处理)")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
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
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "指令执行失败:"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		a.wallet.SendTx(cmdTx)
	} else {
		cmdTx, err := service.NewSendTxCmd(uuID, assetProperty.AssetName, userAddress.PrivateKey,
			params.Address, "", "", params.Amount)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "指令执行失败:"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		a.wallet.SendTx(cmdTx)
	}

	ack := v1.AckWithdrawal{
		OrderID:     uuID,
		UserOrderID: params.UserOrderID,
	}

	pack, err := json.Marshal(ack)
	if err != nil {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "返回数据包错误")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}
	res.Value.Message = string(pack)
	return nil
}

func (a *Address) SupportAssets(req *data.SrvRequest, res *data.SrvResponse) error {
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

func (a *Address) AssetAttribute(req *data.SrvRequest, res *data.SrvResponse) error {
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

	if params.TotalLines == 0 {
		dataList.TotalLines = mysqlpool.QueryAssetPropertyCount(queryMap)
	} else if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
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

func (a *Address) SetAssetAttribute(req *data.SrvRequest, res *data.SrvResponse) error {
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

	params := v1.ReqSetAssetAttribute{
		AssetName:       "",
		FullName:        "",
		IsToken:         -1,
		ParentName:      "",
		DepositMin:      0,
		WithdrawalRate:  0,
		WithdrawalValue: 0,
		ConfirmationNum: 0,
		Decimals:        0,
		Enabled:         -1,
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
	mysqlpool.SetAssetProperty(&assetProperty)
	return nil
}

func (a *Address) GetBalance(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
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

	dataList := v1.AckUserBalanceList{
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
		dataList.TotalLines = mysqlpool.QueryUserAccountCount(queryMap)
	} else if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	}

	if arr, ok := mysqlpool.QueryUserAccount(queryMap); ok {
		for _, v := range arr {
			d := v1.AckUserBalance{}
			d.AssetName = v.AssetName
			d.AvailableAmount = v.AvailableAmount
			d.FrozenAmount = v.FrozenAmount
			d.Time = v.UpdateTime
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

func (a *Address) HistoryTransactionBill(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqHistoryTransactionBill{
		ID:             -1,
		OrderID:        "",
		AssetName:      "",
		Address:        "",
		TransType:      -1,
		Status:         -1,
		Hash:           "",
		MaxAmount:      -1,
		MinAmount:      -1,
		MaxConfirmTime: -1,
		MinConfirmTime: -1,
		TotalLines:     -1,
		PageIndex:      -1,
		MaxDispLines:   -1,
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
	if userProperty.UserClass == 0 {
		queryMap["user_key"] = userProperty.UserKey
	}

	if params.ID > 0 {
		queryMap["id"] = params.ID
	}

	if len(params.OrderID) > 0 {
		queryMap["order_id"] = params.OrderID
	}

	if len(params.AssetName) > 0 {
		queryMap["asset_name"] = params.AssetName
	}

	if len(params.Address) > 0 {
		queryMap["address"] = params.Address
	}

	if params.TransType >= 0 {
		queryMap["trans_type"] = params.TransType
	}

	if params.Status >= 0 {
		queryMap["status"] = params.Status
	}

	if len(params.Hash) > 0 {
		queryMap["hash"] = params.Hash
	}

	if params.MaxAmount >= 0 {
		queryMap["max_amount"] = params.MaxAmount
	}

	if params.MinAmount >= 0 {
		queryMap["min_amount"] = params.MinAmount
	}

	if params.MaxConfirmTime > 0 {
		queryMap["max_confirm_time"] = params.MaxConfirmTime
	}

	if params.MinConfirmTime > 0 {
		queryMap["min_confirm_time"] = params.MinConfirmTime
	}

	dataList := v1.AckHistoryTransactionBillList{
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
		dataList.TotalLines = mysqlpool.QueryTransactionBillCount(queryMap)
	} else if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	}

	if arr, ok := mysqlpool.QueryTransactionBill(queryMap); ok {
		for _, v := range arr {
			d := v1.AckHistoryTransactionBill{}
			d.ID = v.ID
			d.OrderID = v.OrderID
			d.UserOrderID = v.UserOrderID
			d.AssetName = v.AssetName
			d.Address = v.Address
			d.TransType = v.TransType
			d.Amount = v.Amount
			d.PayFee = v.PayFee
			d.Balance = v.Balance
			d.Hash = v.Hash
			d.Status = v.Status
			d.BlockinHeight = v.BlockinHeight
			d.CreateOrderTime = v.CreateOrderTime
			d.BlockinTime = v.BlockinTime
			d.ConfirmTime = v.ConfirmTime
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

func (a *Address) HistoryTransactionMessage(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
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
	if userProperty.UserClass == 0 {
		queryMap["user_key"] = userProperty.UserKey
	}

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
		dataList.TotalLines = mysqlpool.QueryTransactionMessageCount(queryMap)
	} else if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	}

	if arr, ok := mysqlpool.QueryTransactionMessage(queryMap); ok {
		for _, v := range arr {
			d := v1.AckHistoryTransactionMessage{}
			d.MsgID = v.MsgID
			d.TransType = v.TransType
			d.Status = v.Status
			d.BlockinHeight = v.BlockinHeight
			d.AssetName = v.AssetName
			d.Address = v.Address
			d.Amount = v.Amount
			d.PayFee = v.PayFee
			d.Balance = v.Balance
			d.Hash = v.Hash
			d.OrderId = v.OrderID
			d.Time = v.Time
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

func (a *Address) QueryUserAddress(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userProperty.UserKey)
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
	if userProperty.UserClass == 0 {
		queryMap["user_key"] = userProperty.UserKey
	}

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
	} else if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	}

	if arr, ok := mysqlpool.QueryUserAddress(queryMap); ok {
		for _, v := range arr {
			d := v1.AckUserAddress{}
			d.AssetName = v.AssetName
			d.Address = v.Address
			d.AllocationTime = v.AllocationTime

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

func (a *Address) SetPayAddress(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+req.Argv.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if !(userProperty.UserClass == 1 || userProperty.UserClass == 2) {
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
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+req.Argv.UserKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if !(userProperty.UserClass == 1 || userProperty.UserClass == 2) {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "此不能操作此命令")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqPayAddress{}
	if len(req.Argv.Message) > 0 {
		err := json.Unmarshal([]byte(req.Argv.Message), &params)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "解析Json失败-"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
	}

	dataList := v1.AckPayAddressList{}
	if arr, ok := mysqlpool.QueryPayAddressList(params.AssetNames); ok {
		for _, v := range arr {
			d := v1.AckPayAddress{}
			d.AssetName = v.AssetName
			d.Address = v.Address
			d.Amount = v.AvailableAmount
			d.UpdateTime = v.UpdateTime
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

func (a *Address) TransactionBillDaily(req *data.SrvRequest, res *data.SrvResponse) error {
	userKey := req.GetAccessUserKey()
	userProperty, ok := mysqlpool.QueryUserPropertyByKey(userKey)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "无效用户-"+userKey)
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	params := v1.ReqTransactionBillDaily{
		AssetName:    "",
		MaxPeriod:    -1,
		MinPeriod:    -1,
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

	if len(params.AssetName) > 0 {
		queryMap["asset_name"] = params.AssetName
	}

	if params.MaxPeriod >= 0 {
		queryMap["max_period"] = params.MaxPeriod
	}

	if params.MinPeriod >= 0 {
		queryMap["min_period"] = params.MinPeriod
	}

	dataList := v1.AckTransactionBillDailyList{
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
		dataList.TotalLines = mysqlpool.QueryTransactionBillDailyCount(queryMap)
	} else if params.TotalLines > 0 {
		dataList.TotalLines = params.TotalLines
	}

	if arr, ok := mysqlpool.QueryTransactionBillDaily(queryMap); ok {
		for _, v := range arr {
			d := v1.AckTransactionBillDaily{}
			d.Period = v.Period
			d.AssetName = v.AssetName
			d.SumDPAmount = v.SumDPAmount
			d.SumWDAmount = v.SumWDAmount
			d.SumPayFee = v.SumPayFee
			d.PreBalance = v.PreBalance
			d.LastBalance = v.LastBalance
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

func (a *Address) PostTransaction(req *data.SrvRequest, res *data.SrvResponse) error {
	params := backend.ReqPostTransaction{
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

	userAddress, ok := mysqlpool.QueryUserAddressByNameAddress(params.AssetName, params.From)
	if !ok {
		res.Err, res.ErrMsg = CheckError(ErrorFailed, "\"From\"参数无效")
		l4g.Error(res.ErrMsg)
		return errors.New(res.ErrMsg)
	}

	if params.Amount >= userAddress.AvailableAmount {
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
			a.wallet.SendTx(cmdTx)
		} else {
			cmdTx, err := service.NewSendTxCmd("", assetProperty.ParentName, userAddress.PrivateKey,
				params.To, assetProperty.AssetName, userAddress.PrivateKey, params.Amount)
			if err != nil {
				res.Err, res.ErrMsg = CheckError(ErrorFailed, "指令执行失败:"+err.Error())
				l4g.Error(res.ErrMsg)
				return errors.New(res.ErrMsg)
			}
			a.wallet.SendTx(cmdTx)
		}
	} else {
		cmdTx, err := service.NewSendTxCmd("", assetProperty.AssetName, userAddress.PrivateKey,
			params.To, "", "", params.Amount)
		if err != nil {
			res.Err, res.ErrMsg = CheckError(ErrorFailed, "指令执行失败:"+err.Error())
			l4g.Error(res.ErrMsg)
			return errors.New(res.ErrMsg)
		}
		a.wallet.SendTx(cmdTx)
	}
	return nil
}
