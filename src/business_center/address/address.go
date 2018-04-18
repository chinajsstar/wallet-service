package address

import (
	"api_router/base/data"
	"blockchain_server/service"
	"blockchain_server/types"
	"business_center/basicdata"
	. "business_center/def"
	"business_center/mysqlpool"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	for _, v := range basicdata.Get().GetAllUserAddressMap() {
		rcaCmd := service.NewRechargeAddressCmd("", v.AssetName, []string{v.Address})
		a.wallet.InsertRechargeAddress(rcaCmd)
	}
}

func (a *Address) Stop() {
	a.waitGroup.Wait()
}

func (a *Address) NewAddress(req *data.SrvRequestData, res *data.SrvResponseData) error {
	var reqInfo ReqNewAddress
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &reqInfo)
	if err != nil {
		fmt.Printf("NewAddress Unmarshal Error : %s/n", err.Error())
		return err
	}

	var rspInfo RspNewAddress
	rspInfo.ID = reqInfo.ID
	rspInfo.Symbol = reqInfo.Symbol

	userProperty, ok := basicdata.Get().GetAllUserPropertyMap()[req.Data.Argv.LicenseKey]
	if !ok {
		res.Data.Err = -1
		res.Data.ErrMsg = "NewAddress mapUserProperty find Error"
		return errors.New(res.Data.ErrMsg)
	}

	assetProperty, ok := basicdata.Get().GetAllAssetPropertyMap()[reqInfo.Symbol]
	if !ok {
		res.Data.Err = -1
		res.Data.ErrMsg = "NewAddress mapAssetProperty find Error"
		return errors.New(res.Data.ErrMsg)
	}

	userAddresses := a.generateAddress(userProperty.UserKey, userProperty.UserClass, assetProperty.ID,
		assetProperty.Name, reqInfo.Count)
	if len(userAddresses) > 0 {
		rspInfo.Address = a.addUserAddress(userAddresses)

		strTime := time.Now().UTC().Format("2006-01-02 15:04:05")
		db := mysqlpool.Get()
		db.Exec("insert user_account (user_id, asset_id, available_amount, frozen_amount,"+
			" create_time, update_time) values (?, ?, 0, 0, ?, ?);",
			userProperty.UserKey, assetProperty.ID,
			strTime, strTime)

		//添加监控地址
		rcaCmd := service.NewRechargeAddressCmd("message id", assetProperty.Name, rspInfo.Address)
		a.wallet.InsertRechargeAddress(rcaCmd)
	}

	pack, err := json.Marshal(rspInfo)
	if err != nil {
		res.Data.Err = -1
		res.Data.ErrMsg = "NewAddress RspNewAddress Marshal Error"
		fmt.Printf("NewAddress RspNewAddress Marshal Error : %s/n", err.Error())
		return err
	}

	res.Data.Value.Message = string(pack)
	res.Data.Err = 0
	res.Data.ErrMsg = ""

	return nil
}

func (a *Address) Withdrawal(req *data.SrvRequestData, res *data.SrvResponseData) error {
	var reqInfo ReqWithdrawal
	err := json.Unmarshal([]byte(req.Data.Argv.Message), &reqInfo)
	if err != nil {
		fmt.Printf("Withdrawal Unmarshal Error : %s/n", err.Error())
		return err
	}

	userProperty, ok := basicdata.Get().GetAllUserPropertyMap()[req.Data.Argv.LicenseKey]
	if !ok {
		return errors.New("withdrawal mapUserProperty find Error")
	}

	assetProperty, ok := basicdata.Get().GetAllAssetPropertyMap()[reqInfo.Symbol]
	if !ok {
		return errors.New("withdrawal mapAssetProperty find Error")
	}

	var rspInfo RspWithdrawal
	uID, _ := uuid.NewV4()
	rspInfo.OrderID = uID.String()
	rspInfo.UserOrderID = reqInfo.UserOrderID
	rspInfo.Timestamp = time.Now().Unix()

	value := int64(reqInfo.Amount * math.Pow10(18) * assetProperty.WithdrawalRate)

	Tx, err := mysqlpool.Get().Begin()
	if err != nil {
		return err
	}

	ret, err := Tx.Exec("update user_account set available_amount = available_amount - ?, frozen_amount = frozen_amount + ?,"+
		" update_time = ? where user_id = ? and asset_id = ? and available_amount >= ?;",
		value, value,
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		userProperty.UserKey,
		assetProperty.ID,
		value)

	if err != nil {
		Tx.Rollback()
		return err
	}

	rows, _ := ret.RowsAffected()
	if rows < 1 {
		Tx.Rollback()
		return nil
	}

	ret, err = Tx.Exec("update user_address a set"+
		" a.available_amount = a.available_amount - ?,"+
		" a.frozen_amount = a.frozen_amount + ?"+
		" where a.available_amount >= ? and (a.asset_id, a.address) in (select asset_id, address from pay_address)",
		value, value, value)

	if err != nil {
		Tx.Rollback()
		return err
	}

	rows, _ = ret.RowsAffected()
	if rows < 1 {
		Tx.Rollback()
		return nil
	}

	_, err = Tx.Exec("insert withdraw_order (order_id, user_order_id, user_id, asset_id, address, amount, wallet_fee, create_time) "+
		"values (?, ?, ?, ?, ?, ?, ?, ?);",
		rspInfo.OrderID, reqInfo.UserOrderID, userProperty.UserKey, assetProperty.ID,
		reqInfo.ToAddress, int64(reqInfo.Amount*math.Pow10(18)), 0,
		time.Now().UTC().Format("2006-01-02 15:04:05"))

	Tx.Commit()

	pack, err := json.Marshal(rspInfo)
	if err != nil {
		fmt.Printf("withdrawal RspNewAddress Marshal Error : %s/n", err.Error())
		return err
	}

	//txCmd := service.NewSendTxCmd("message id", coin, privatekey, to, token, value)
	//a.wallet.SendTx()

	res.Data.Value.Message = string(pack)
	res.Data.Err = 0
	res.Data.ErrMsg = ""

	return nil
}

func (a *Address) QueryUserAddress(req *data.SrvRequestData, res *data.SrvResponseData) error {
	mapUserAddress, err := mysqlpool.QueryAllUserAddress()
	if err != nil {
		return err
	}

	addresses := make([]UserAddress, 0)
	for _, v := range mapUserAddress {
		addresses = append(addresses, *v)
	}

	pack, _ := json.Marshal(addresses)
	res.Data.Value.Message = string(pack)
	res.Data.Err = 0
	res.Data.ErrMsg = ""

	return nil
}
