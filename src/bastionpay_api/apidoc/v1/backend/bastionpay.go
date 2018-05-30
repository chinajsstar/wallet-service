package backend

import (
	"bastionpay_api/api/v1"
	"bastionpay_api/apibackend"
	"bastionpay_api/apibackend/v1/backend"
	"bastionpay_api/apidoc"
)

var ApiDocSpPostTransaction = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterAdmin},
	Name:        "发送交易",
	Description: "自定义发送交易，管理员权限api",
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "sp_post_transaction",
	Input:       &backend.SpReqPostTransaction{AssetName: "BTC", From: "mfetgXxCwsh9U19v3Kv5j77FdgbchHjkzK", To: "mrEfgUBMUM5zjmzSdoBQuodTz16kyZ1tnD", Amount: 1.0},
	Output:      new(string),
}

var ApiDocSpReqAssetsAttributeList = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterAdmin},
	Name:        "查询币种属性",
	Description: "查询币种属性，管理员权限api",
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "sp_get_asset_attribute",
	Input:       &backend.SpReqAssetsAttributeList{},
	Output:      new(string),
}

var ApiDocSetPayAddress = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterAdmin},
	Name:        "设置热钱包地址",
	Description: "设置热钱包地址，管理员权限api",
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "set_pay_address",
	Input:       new(string),
	Output:      new(string),
}

// 设置币种属性
var ApiDocSetAssetAttribute = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterAdmin},
	Name:        "设置币种属性",
	Description: "设置币种属性",
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "set_asset_attribute",
	Input:       &v1.ReqSetAssetAttribute{},
	Output:      new(string),
}
