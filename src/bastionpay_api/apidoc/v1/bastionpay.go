package v1

import (
	"bastionpay_api/api/v1"
	"bastionpay_api/apibackend"
	"bastionpay_api/apidoc"
)

// 支持币种
var ApiDocSupportAssets = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "support_assets",
	Input:       &v1.ReqSupportAssets{},
	Output:      &v1.AckSupportAssetList{Data: []string{"btc"}},
}

// 币种属性
var ApiDocAssetAttribute = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "asset_attribute",
	Input:       &v1.ReqAssetsAttributeList{AssetNames: []string{"btc"}},
	Output:      &v1.AckAssetsAttributeList{Data: []v1.AckAssetsAttribute{v1.AckAssetsAttribute{}}},
}

// 币种余额
var ApiDocGetBalance = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "get_balance",
	Input:       &v1.ReqUserBalance{AssetNames: []string{"btc"}},
	Output:      &v1.AckUserBalanceList{Data: []v1.AckUserBalance{v1.AckUserBalance{}}},
}

// 用户地址
var ApiDocQueryAddress = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "query_address",
	Input:       &v1.ReqUserAddress{},
	Output:      &v1.AckUserAddressList{Data: []v1.AckUserAddress{v1.AckUserAddress{}}},
}

// 历史交易订单
var ApiDocTransactionBill = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "transaction_bill",
	Input:       &v1.ReqHistoryTransactionBill{},
	Output:      &v1.AckHistoryTransactionBillList{Data: []v1.AckHistoryTransactionBill{v1.AckHistoryTransactionBill{}}},
}

// 日结帐单查询
var ApiDocTransactionBillDaily = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "transaction_bill_daily",
	Input:       &v1.ReqTransactionBillDaily{},
	Output:      &v1.AckTransactionBillDailyList{Data: []v1.AckTransactionBillDaily{v1.AckTransactionBillDaily{}}},
}

// 历史交易消息
var ApiDocTransactionMessage = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterApi, apibackend.HttpRouterUser, apibackend.HttpRouterAdmin},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "transaction_message",
	Input:       &v1.ReqHistoryTransactionMessage{},
	Output:      &v1.AckHistoryTransactionMessageList{Data: []v1.AckHistoryTransactionMessage{v1.AckHistoryTransactionMessage{}}},
}

// TODO:以下需继续
var ApiDocNewAddress = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterApi, apibackend.HttpRouterAdmin},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "new_address",
	Input:       new(string),
	Output:      new(string),
}

var ApiDocWithdrawal = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterApi, apibackend.HttpRouterAdmin},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "withdrawal",
	Input:       new(string),
	Output:      new(string),
}
