package v1

import (
	"api_router/base/data"
	"bastionpay_api/api/v1"
	"bastionpay_api/apidoc"
)

// 支持币种
var ApiDocSupportAssets = apidoc.ApiDoc{
	Name:        "获取币种",
	Description: "获取支持币种",
	Level:       data.APILevel_client,
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "support_assets",
	Input:       &v1.ReqSupportAssets{},
	Output:      &v1.AckSupportAssetList{Data: []string{"btc"}},
}

// 币种属性
var ApiDocAssetAttribute = apidoc.ApiDoc{
	Name:        "获取币种属性",
	Description: "获取币种属性",
	Level:       data.APILevel_client,
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "asset_attribute",
	Input:       &v1.ReqAssetsAttributeList{AssetNames: []string{"btc"}},
	Output:      &v1.AckAssetsAttributeList{Data: []v1.AckAssetsAttribute{v1.AckAssetsAttribute{}}},
}

// 币种余额
var ApiDocGetBalance = apidoc.ApiDoc{
	Name:        "查询用户币种余额",
	Description: "查询币种余额",
	Level:       data.APILevel_client,
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "get_balance",
	Input:       &v1.ReqUserBalance{AssetNames: []string{"btc"}},
	Output:      &v1.AckUserBalanceList{Data: []v1.AckUserBalance{v1.AckUserBalance{}}},
}

// 用户地址
var ApiDocQueryUserAddress = apidoc.ApiDoc{
	Name:        "查询用户地址",
	Description: "用户查询地址",
	Level:       data.APILevel_client,
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "query_user_address",
	Input:       &v1.ReqUserAddress{},
	Output:      &v1.AckUserAddressList{Data: []v1.AckUserAddress{v1.AckUserAddress{}}},
}

// 历史交易订单
var ApiDocHistoryTransactionOrder = apidoc.ApiDoc{
	Name:        "查询历史交易订单",
	Description: "查询历史交易订单",
	Level:       data.APILevel_client,
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "history_transaction_order",
	Input:       &v1.ReqHistoryTransactionOrder{},
	Output:      &v1.AckHistoryTransactionOrderList{Data: []v1.AckHistoryTransactionOrder{v1.AckHistoryTransactionOrder{}}},
}

// 历史交易消息
var ApiDocHistoryTransactionMessage = apidoc.ApiDoc{
	Name:        "查询历史交易信息",
	Description: "查询历史交易信息",
	Level:       data.APILevel_client,
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "history_transaction_message",
	Input:       &v1.ReqHistoryTransactionMessage{},
	Output:      &v1.AckHistoryTransactionMessageList{Data: []v1.AckHistoryTransactionMessage{v1.AckHistoryTransactionMessage{}}},
}

// TODO:以下需继续
var ApiDocNewAddress = apidoc.ApiDoc{
	Name:        "生成地址",
	Description: "生成地址",
	Level:       data.APILevel_client,
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "new_address",
	Input:       new(string),
	Output:      new(string),
}

var ApiDocWithdrawal = apidoc.ApiDoc{
	Name:        "提币",
	Description: "提币",
	Level:       data.APILevel_client,
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "withdrawal",
	Input:       new(string),
	Output:      new(string),
}

var ApiDocSetPayAddress = apidoc.ApiDoc{
	Name:        "设置热钱包地址",
	Description: "设置热钱包地址，管理员权限api",
	Level:       data.APILevel_admin,
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "set_pay_address",
	Input:       new(string),
	Output:      new(string),
}
