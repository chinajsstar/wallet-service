package address

import (
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
	rechargeChannel types.RechargeTxChannel
	cmdTxChannel    types.CmdTxChannel
	waitGroup       sync.WaitGroup
	ctx             context.Context
}

func (a *Address) Run(ctx context.Context, wallet *service.ClientManager) {
	a.wallet = wallet
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

func (a *Address) AllocationAddress(req string, ack *string) error {
	var reqInfo ReqNewAddress
	err := json.Unmarshal([]byte(req), &reqInfo)
	if err != nil {
		fmt.Printf("AllocationAddress Unmarshal Error : %s/n", err.Error())
		return err
	}

	var rspInfo RspNewAddress
	rspInfo.Result.ID = reqInfo.UserID
	rspInfo.Result.Symbol = reqInfo.Params.Symbol
	rspInfo.Status.Code = 0
	rspInfo.Status.Msg = ""

	userProperty, ok := basicdata.Get().GetAllUserPropertyMap()[reqInfo.UserID]
	if !ok {
		return errors.New("AllocationAddress mapUserProperty find Error")
	}

	assetProperty, ok := basicdata.Get().GetAllAssetPropertyMap()[reqInfo.Params.Symbol]
	if !ok {
		return errors.New("AllocationAddress mapAssetProperty find Error")
	}

	userAddresses := a.generateAddress(userProperty.UserID, userProperty.UserClass, assetProperty.ID,
		assetProperty.Name, reqInfo.Params.Count)
	if len(userAddresses) > 0 {
		rspInfo.Result.Address = a.addUserAddress(userAddresses)

		strTime := time.Now().UTC().Format("2006-01-02 15:04:05")
		db := mysqlpool.Get()
		db.Exec("insert user_account (user_id, asset_id, available_amount, frozen_amount,"+
			" create_time, update_time) values (?, ?, 0, 0, ?, ?);",
			userProperty.UserID, assetProperty.ID,
			strTime, strTime)

		//添加监控地址
		rcaCmd := service.NewRechargeAddressCmd("message id", assetProperty.Name, rspInfo.Result.Address)
		a.wallet.InsertRechargeAddress(rcaCmd)
	}

	pack, err := json.Marshal(rspInfo)
	if err != nil {
		fmt.Printf("AllocationAddress RspNewAddress Marshal Error : %s/n", err.Error())
		return err
	}
	*ack = string(pack)
	return nil
}

func (a *Address) Withdrawal(req string, ack *string) error {
	var reqInfo ReqWithdrawal
	err := json.Unmarshal([]byte(req), &reqInfo)
	if err != nil {
		fmt.Printf("Withdrawal Unmarshal Error : %s/n", err.Error())
		return err
	}

	userProperty, ok := basicdata.Get().GetAllUserPropertyMap()[reqInfo.UserID]
	if !ok {
		return errors.New("withdrawal mapUserProperty find Error")
	}

	assetProperty, ok := basicdata.Get().GetAllAssetPropertyMap()[reqInfo.Params.Symbol]
	if !ok {
		return errors.New("withdrawal mapAssetProperty find Error")
	}

	value := int64(reqInfo.Params.Amount * math.Pow10(18))
	db := mysqlpool.Get()
	ret, err := db.Exec("update user_account set available_amount = available_amount - ?, frozen_amount = frozen_amount + ?,"+
		" update_time = ? where user_id = ? and asset_id = ? and available_amount >= ?;",
		value, value,
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		userProperty.UserID,
		assetProperty.ID,
		value)

	if err != nil {
		return err
	}

	var rspInfo RspWithdrawal
	uID, _ := uuid.NewV4()
	rspInfo.Result.OrderID = uID.String()
	rspInfo.Result.UserOrderID = reqInfo.Params.UserOrderID
	rspInfo.Result.Timestamp = time.Now().Unix()
	rspInfo.Status.Code = 0
	rspInfo.Status.Msg = ""

	if rows, _ := ret.RowsAffected(); rows > 0 {
		_, err = db.Exec("insert withdraw_order (order_id, user_order_id, user_id, asset_id, address, amount, wallet_fee, create_time) "+
			"values (?, ?, ?, ?, ?, ?, ?, ?);",
			rspInfo.Result.OrderID, reqInfo.Params.UserOrderID, userProperty.UserID, assetProperty.ID,
			reqInfo.Params.ToAddress, int64(reqInfo.Params.Amount*math.Pow10(18)), 0,
			time.Now().UTC().Format("2006-01-02 15:04:05"))
	}

	pack, err := json.Marshal(rspInfo)
	if err != nil {
		fmt.Printf("withdrawal RspNewAddress Marshal Error : %s/n", err.Error())
		return err
	}
	*ack = string(pack)

	//txCmd := service.NewSendTxCmd("message id", coin, privatekey, to, token, value)
	//a.wallet.SendTx()

	return nil
}
