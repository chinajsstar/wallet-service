package v1

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/api/v1"
	"bastionpay_api/apibackend"
)

var ApiDocRecharge = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterAdmin},
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"recharge",
	Input:&v1.ReqRecharge{},
	Output:new(string),
}

var ApiDocGenerate = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterAdmin},
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"generate",
	Input:&v1.ReqGenerate{},
	Output:new(string),
}