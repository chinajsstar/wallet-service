package backend

import (
	"bastionpay_api/apibackend"
	"bastionpay_api/apibackend/v1/backend"
	"bastionpay_api/apidoc"
)

var ApiDocSpPostTransaction = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterAdmin},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "sp_post_transaction",
	Input:       &backend.SpReqPostTransaction{AssetName: "BTC", From: "mfetgXxCwsh9U19v3Kv5j77FdgbchHjkzK", To: "mrEfgUBMUM5zjmzSdoBQuodTz16kyZ1tnD", Amount: 1.0},
	Output:      new(string),
}

var ApiDocSpReqAssetsAttributeList = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterAdmin},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "sp_get_asset_attribute",
	Input:       &backend.SpReqAssetsAttributeList{},
	Output:      &backend.SpAckAssetsAttributeList{},
}

var ApiDocSetPayAddress = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterAdmin},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "sp_set_pay_address",
	Input:       new(string),
	Output:      new(string),
}

// 设置币种属性
var ApiDocSetAssetAttribute = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterAdmin},
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "sp_set_asset_attribute",
	Input:       &backend.SpReqSetAssetAttribute{},
	Output:      new(string),
}
