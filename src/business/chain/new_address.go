package chain

import (
	"api_router/base/data"
	"bastionpay_api/api/v1"
	"blockchain_server/service"
	. "business/def"
	"business/mysqlpool"
	"encoding/json"
	"errors"
	l4g "github.com/alecthomas/log4go"
	"time"
)

func NewAddress(wallet *service.ClientManager, req *data.SrvRequest, res *data.SrvResponse) error {
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

	userAddress := generateAddress(wallet, &userProperty, &assetProperty, params.Count)
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

func generateAddress(wallet *service.ClientManager, userProperty *UserProperty, assetProperty *AssetProperty, count int) []UserAddress {
	assetName := assetProperty.AssetName
	if assetProperty.IsToken > 0 {
		assetName = assetProperty.ParentName
	}
	cmd := service.NewAccountCmd("", assetName, 1)
	userAddress := make([]UserAddress, 0)
	for i := 0; i < count; i++ {
		accounts, err := wallet.NewAccounts(cmd)
		if err != nil {
			CheckError(ErrorFailed, err.Error())
			return []UserAddress{}
		}
		nowTM := time.Now().Unix()
		data := UserAddress{
			UserKey:         userProperty.UserKey,
			UserClass:       userProperty.UserClass,
			AssetName:       assetProperty.AssetName,
			Address:         accounts[0].Address,
			PrivateKey:      accounts[0].PrivateKey,
			AvailableAmount: 0,
			FrozenAmount:    0,
			Enabled:         1,
			CreateTime:      nowTM,
			AllocationTime:  nowTM,
			UpdateTime:      nowTM,
		}

		//添加地址监控
		cmd := service.NewRechargeAddressCmd("", assetName, []string{data.Address})
		err = wallet.InsertRechargeAddress(cmd)
		if err != nil {
			CheckError(ErrorFailed, err.Error())
			return []UserAddress{}
		}
		userAddress = append(userAddress, data)
	}
	err := mysqlpool.AddUserAddress(userAddress)
	if err != nil {
		return []UserAddress{}
	}

	if userProperty.UserClass == 0 {
		err = mysqlpool.AddUserAccount(userProperty.UserKey, userProperty.UserClass, assetProperty.AssetName)
		if err != nil {
			return []UserAddress{}
		}
	}

	if userProperty.UserClass == 1 && assetProperty.IsToken == 0 {
		mysqlpool.CreateTokenAddress(userAddress)
	}

	return userAddress
}
