package v1

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/api/v1"
	"bastionpay_api/apibackend"
)

var ApiDocAuth = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterUser,apibackend.HttpRouterAdmin},
	Name:"验证解密",
	Description:"验证解密数据",
	VerName:"v1",
	SrvName:"auth",
	FuncName:"authdata",
	Input:&v1.ReqAuth{},
	Output:new(string),
}

var ApiDocEncrypt = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterUser,apibackend.HttpRouterAdmin},
	Name:"加密签名",
	Description:"加密签名数据",
	VerName:"v1",
	SrvName:"auth",
	FuncName:"encryptdata",
	Input:new(string),
	Output:new(string),
}