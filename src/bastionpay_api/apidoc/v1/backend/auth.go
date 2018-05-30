package backend

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/apibackend"
)

var ApiDocAuth = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterUser,apibackend.HttpRouterAdmin},
	VerName:"v1",
	SrvName:"auth",
	FuncName:"authdata",
	Input:new(string),
	Output:new(string),
}

var ApiDocEncrypt = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterUser,apibackend.HttpRouterAdmin},
	VerName:"v1",
	SrvName:"auth",
	FuncName:"encryptdata",
	Input:new(string),
	Output:new(string),
}