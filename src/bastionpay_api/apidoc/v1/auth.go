package v1

import (
	"bastionpay_api/apidoc"
	"api_router/base/data"
	"bastionpay_api/api/v1"
)

var ApiDocAuth = apidoc.ApiDoc{
	Name:"验证解密",
	Description:"验证解密数据",
	Level:data.APILevel_client,
	VerName:"v1",
	SrvName:"auth",
	FuncName:"authdata",
	Input:&v1.ReqAuth{},
	Output:new(string),
}

var ApiDocEncrypt = apidoc.ApiDoc{
	Name:"加密签名",
	Description:"加密签名数据",
	Level:data.APILevel_client,
	VerName:"v1",
	SrvName:"auth",
	FuncName:"encryptdata",
	Input:new(string),
	Output:new(string),
}