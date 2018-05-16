package v1

import (
	"bastionpay_api/apidoc"
	"api_router/base/data"
	"bastionpay_api/api/v1"
)

var ApiDocAuth = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"auth",
	FuncName:"authdata",
	Level:data.APILevel_client,
	Comment:"验证解密数据",
	Input:&v1.ReqAuth{},
	Output:new(string),
}

var ApiDocEncrypt = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"auth",
	FuncName:"encryptdata",
	Level:data.APILevel_client,
	Comment:"加密签名数据",
	Input:new(string),
	Output:new(string),
}