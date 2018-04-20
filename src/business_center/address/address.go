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

	userProperty, ok := basicdata.Get().GetAllUserPropertyMap()[req.Data.Argv.UserKey]
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
		db.Exec("insert user_account (user_key, asset_id, available_amount, frozen_amount,"+
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

	var rspInfo RspWithdrawal
	rspInfo.UserOrderID = reqInfo.UserOrderID
	rspInfo.Timestamp = time.Now().Unix()
	res.Data.Err = 0
	res.Data.ErrMsg = ""

	userProperty, ok := basicdata.Get().GetAllUserPropertyMap()[req.Data.Argv.UserKey]
	if !ok {
		return errors.New("withdrawal mapUserProperty find Error")
	}

	assetProperty, ok := basicdata.Get().GetAllAssetPropertyMap()[reqInfo.Symbol]
	if !ok {
		return errors.New("withdrawal mapAssetProperty find Error")
	}

	row := mysqlpool.Get().QueryRow("select a.address, a.private_key,"+
		" b.available_amount, b.frozen_amount from pay_address a"+
		" left join user_address b on a.asset_id = b.asset_id and a.address = b.address"+
		" where a.asset_id = ?", assetProperty.ID)
	if row == nil {
		return nil
	}

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
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		userProperty.UserKey,
		assetProperty.ID,
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

	_, err = Tx.Exec("insert withdraw_order (order_id, user_order_id, user_key, asset_id, address, amount, wallet_fee, create_time) "+
		"values (?, ?, ?, ?, ?, ?, ?, ?);",
		rspInfo.OrderID, reqInfo.UserOrderID, userProperty.UserKey, assetProperty.ID,
		reqInfo.ToAddress, amount, fee,
		time.Now().UTC().Format("2006-01-02 15:04:05"))

	Tx.Commit()

	pack, err := json.Marshal(rspInfo)
	if err != nil {
		fmt.Printf("withdrawal RspNewAddress Marshal Error : %s/n", err.Error())
		return err
	}

	txCmd := service.NewSendTxCmd(rspInfo.OrderID, assetProperty.Name, privateKey, reqInfo.ToAddress, nil, uint64(amount))
	a.wallet.SendTx(txCmd)
	res.Data.Value.Message = string(pack)

	return nil
}

func (a *Address) QueryUserAddress(req *data.SrvRequestData, res *data.SrvResponseData) error {
	res.Data.Value.Message = mysqlpool.QueryUserAddress(req.Data.Argv.Message)
	res.Data.Err = 0
	res.Data.ErrMsg = ""
	return nil
}
