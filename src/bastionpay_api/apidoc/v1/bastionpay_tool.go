package v1

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/api/v1"
	"api_router/base/data"
)

var ApiDocRecharge = apidoc.ApiDoc{
	Name:"模拟充值",
	Description:"模拟充值",
	Level:data.APILevel_client,
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"recharge",
	Input:&v1.ReqRecharge{},
	Output:new(string),
}

var ApiDocGenerate = apidoc.ApiDoc{
	Name:"模拟挖矿",
	Description:"模拟挖矿",
	Level:data.APILevel_client,
	VerName:"v1",
	SrvName:"bastionpay",
	FuncName:"generate",
	Input:&v1.ReqGenerate{},
	Output:new(string),
}