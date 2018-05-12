package v1

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/api/v1"
	"api_router/base/data"
)

var ApiDocSupportAssets = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"support_assets",
	Level:data.APILevel_client,
	Comment:"获取支持币种",
	Path:"/api/v1/bastionpay/support_assets",
	Input:nil,
	Output:[]string{"btc"},
}

var ApiDocAssetAttribute = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"asset_attribute",
	Level:data.APILevel_client,
	Comment:"获取币种属性",
	Path:"/api/v1/bastionpay/asset_attribute",
	Input:[]string{"btc"},
	Output:v1.AckAssetsAttributes{v1.AckAssetsAttribute{}},
}

var ApiDocNewAddress = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"new_address",
	Level:data.APILevel_client,
	Comment:"生成地址",
	Path:"/api/v1/bastionpay/new_address",
	Input:nil,
	Output:nil,
}

var ApiDocQueryUserAddress = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"query_user_address",
	Level:data.APILevel_client,
	Comment:"用户查询地址",
	Path:"/api/v1/bastionpay/query_user_address",
	Input:nil,
	Output:nil,
}

var ApiDocWithdrawal = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"withdrawal",
	Level:data.APILevel_client,
	Comment:"提币",
	Path:"/api/v1/bastionpay/withdrawal",
	Input:nil,
	Output:nil,
}

var ApiDocGetBalance = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"get_balance",
	Level:data.APILevel_client,
	Comment:"查询币种余额",
	Path:"/api/v1/bastionpay/get_balance",
	Input:nil,
	Output:nil,
}

var ApiDocHistoryTransactionOrder = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"history_transaction_order",
	Level:data.APILevel_client,
	Comment:"查询历史交易订单",
	Path:"/api/v1/bastionpay/history_transaction_order",
	Input:nil,
	Output:nil,
}

var ApiDocHistoryTransactionMessage = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"history_transaction_message",
	Level:data.APILevel_client,
	Comment:"查询历史交易信息",
	Path:"/api/v1/bastionpay/history_transaction_message",
	Input:nil,
	Output:nil,
}

var ApiDocSetPayAddress = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"set_pay_address",
	Level:data.APILevel_client,
	Comment:"设置热钱包地址",
	Path:"/api/v1/bastionpay/set_pay_address",
	Input:nil,
	Output:nil,
}
