package backend

import (
	"bastionpay_api/apibackend"
	"bastionpay_api/apibackend/v1/backend"
	"bastionpay_api/apidoc"
)

var ApiDocPostTransaction = apidoc.ApiDoc{
	Group:       []string{apibackend.HttpRouterAdmin},
	Name:        "发送交易",
	Description: "自定义发送交易，管理员权限api",
	VerName:     "v1",
	SrvName:     "bastionpay",
	FuncName:    "sp_post_transaction",
	Input:       &backend.ReqPostTransaction{AssetName: "BTC", From: "mfetgXxCwsh9U19v3Kv5j77FdgbchHjkzK", To: "mrEfgUBMUM5zjmzSdoBQuodTz16kyZ1tnD", Amount: 1.0},
	Output:      new(string),
}
