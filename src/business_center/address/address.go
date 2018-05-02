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
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"github.com/satori/go.uuid"
	"math"
	"sync"
	"time"
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
			if assetProperty, ok := mysqlpool.QueryAssetPropertyByID(v.AssetID); ok {
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
	resMap["user_order_id"] = paramsMapping.UserOrderID
	resMap["asset_id"] = paramsMapping.AssetID

	assetProperty, ok := mysqlpool.QueryAssetPropertyByID(paramsMapping.AssetID)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"asset_id\"无效")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	if paramsMapping.Count <= 0 {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "参数:\"count\"要大于0")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}
	resMap["data"] = a.generateAddress(&userProperty, &assetProperty, paramsMapping.Count)
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
	resMap["user_order_id"] = paramsMapping.UserOrderID
	resMap["asset_id"] = paramsMapping.AssetID

	assetProperty, ok := mysqlpool.QueryAssetPropertyByID(paramsMapping.AssetID)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "获取用户帐户信息失败")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	userAccount, ok := mysqlpool.QueryUserAccountByKey(userProperty.UserKey)
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

	walletFee := assetProperty.WithdrawalValue + int64(float64(paramsMapping.Amount)*assetProperty.WithdrawalRate)
	if userAccount.AvailableAmount < paramsMapping.Amount+walletFee {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorParse, "用户帐户可用资金不足")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	userAddress, ok := mysqlpool.QueryPayAddress(assetProperty.AssetID)
	if !ok {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorWallet, "没有设置可用的热钱包")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	if userAddress.AvailableAmount < paramsMapping.Amount+walletFee {
		res.Data.Err, res.Data.ErrMsg = CheckError(ErrorWallet, "热钱包资金不足，这里需要特殊处理")
		l4g.Error(res.Data.ErrMsg)
		return errors.New(res.Data.ErrMsg)
	}

	uuID := a.generateUUID()
	resMap["order_id"] = uuID

	err := mysqlpool.WithDrawalSet(userProperty.UserKey, assetProperty.AssetID, paramsMapping.Address,
		paramsMapping.Amount, walletFee, uuID)
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
		a.wallet.SendTx(service.NewSendTxCmd(paramsMapping.UserOrderID, assetProperty.CoinName, userAddress.PrivateKey,
			paramsMapping.Address, &assetProperty.AssetName, uint64(paramsMapping.Amount)))
	} else {
		a.wallet.SendTx(service.NewSendTxCmd(paramsMapping.UserOrderID, assetProperty.AssetName, userAddress.PrivateKey,
			paramsMapping.Address, nil, uint64(paramsMapping.Amount)))
	}
	res.Data.Value.Message = string(pack)
	return nil
}

func (a *Address) oldWithdrawal(req *data.SrvRequestData, res *data.SrvResponseData) error {
	var reqInfo ReqWithdrawal
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &reqInfo)
	if err != nil {
		fmt.Printf("Withdrawal Unmarshal Error : %s/n", err.Error())
		return err
	}

	var rspInfo RspWithdrawal
	rspInfo.UserOrderID = reqInfo.UserOrderID
	rspInfo.Timestamp = time.Now().Unix()
	res.Data.Err = 0
	res.Data.ErrMsg = ""

	userProperty, ok := mysqlpool.QueryUserPropertyByKey(req.Data.Argv.UserKey)
	if !ok {
		err := errors.New("withdrawal UserProperty Find Error")
		res.Data.Err = -1
		res.Data.ErrMsg = err.Error()
		return err
	}

	assetProperty, ok := mysqlpool.QueryAssetPropertyByName(reqInfo.Symbol)
	if !ok {
		err := errors.New("withdrawal AssetProperty Find Error")
		res.Data.Err = -1
		res.Data.ErrMsg = err.Error()
		return err
	}

	row := mysqlpool.Get().QueryRow("select a.address, a.private_key,"+
		" b.available_amount, b.frozen_amount from pay_address a"+
		" left join user_address b on a.asset_id = b.asset_id and a.address = b.address"+
		" where a.asset_id = ?", assetProperty.AssetID)

	var (
		address         string
		privateKey      string
		availableAmount int64
		frozenAmount    int64
	)
	err = row.Scan(&address, &privateKey, &availableAmount, &frozenAmount)
	if err != nil {
		fmt.Println("没有设置热钱包")
		return err
	}

	amount := int64(reqInfo.Amount * math.Pow10(8))
	fee := int64(float64(amount) * assetProperty.WithdrawalRate)

	if availableAmount < amount+fee {
		fmt.Println("热钱包资金不够")
		return nil
	}

	Tx, err := mysqlpool.Get().Begin()
	if err != nil {
		return err
	}

	ret, err := Tx.Exec("update user_account set available_amount = available_amount - ?, frozen_amount = frozen_amount + ?,"+
		" update_time = ? where user_key = ? and asset_id = ? and available_amount >= ?;",
		amount+fee, amount+fee,
		time.Now().UTC().Format(TimeFormat),
		userProperty.UserKey,
		assetProperty.AssetID,
		amount+fee)

	if err != nil {
		Tx.Rollback()
		return err
	}

	rows, _ := ret.RowsAffected()
	if rows < 1 {
		Tx.Rollback()
		return nil
	}

	uID, _ := uuid.NewV4()
	rspInfo.OrderID = uID.String()

	_, err = Tx.Exec("insert withdrawal_order (order_id, user_order_id, user_key, asset_id, address, amount, wallet_fee, create_time) "+
		"values (?, ?, ?, ?, ?, ?, ?, ?);",
		rspInfo.OrderID, reqInfo.UserOrderID, userProperty.UserKey, assetProperty.AssetID,
		reqInfo.ToAddress, amount, fee,
		time.Now().UTC().Format(TimeFormat))

	Tx.Commit()

	pack, err := json.Marshal(rspInfo)
	if err != nil {
		fmt.Printf("withdrawal RspNewAddress Marshal Error : %s/n", err.Error())
		return err
	}

	if assetProperty.IsToken > 0 {
		txCmd := service.NewSendTxCmd(rspInfo.OrderID, assetProperty.CoinName, privateKey, reqInfo.ToAddress, &assetProperty.AssetName, uint64(amount))
		a.wallet.SendTx(txCmd)
	} else {
		txCmd := service.NewSendTxCmd(rspInfo.OrderID, assetProperty.AssetName, privateKey, reqInfo.ToAddress, nil, uint64(amount))
		a.wallet.SendTx(txCmd)
	}
	res.Data.Value.Message = string(pack)

	return nil
}

func (a *Address) QueryAssetProperty(req *data.SrvRequestData, res *data.SrvResponseData) error {
	query := req.Data.Argv.Message
	resMap := responsePagination(query, mysqlpool.QueryAssetPropertyCount(query))
	assetProperty, _ := mysqlpool.QueryAssetProperty(query)
	resMap["data"] = assetProperty

	res.Data.Value.Message = packJson(resMap)
	res.Data.Err = 0
	res.Data.ErrMsg = ""
	return nil
}

func (a *Address) QueryUserProperty(req *data.SrvRequestData, res *data.SrvResponseData) error {
	query := req.Data.Argv.Message
	resMap := responsePagination(query, mysqlpool.QueryUserPropertyCount(query))
	userProperty, _ := mysqlpool.QueryUserProperty(query)
	resMap["data"] = userProperty

	res.Data.Value.Message = packJson(resMap)
	res.Data.Err = 0
	res.Data.ErrMsg = ""
	return nil
}

func (a *Address) QueryUserAccount(req *data.SrvRequestData, res *data.SrvResponseData) error {
	query := req.Data.Argv.Message
	resMap := responsePagination(query, mysqlpool.QueryUserAccountCount(query))
	userAccount, _ := mysqlpool.QueryUserAccount(query)
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
