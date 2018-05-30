package backend

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/apibackend"
	"bastionpay_api/apibackend/v1/backend"
)

var ApiDocRegister = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterUser,apibackend.HttpRouterAdmin},
	VerName:"v1",
	SrvName:"account",
	FuncName:"register",
	Input:&backend.ReqUserRegister{},
	Output:&backend.AckUserRegister{},
}

var ApiDocUpdateProfile = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterUser,apibackend.HttpRouterAdmin},
	VerName:"v1",
	SrvName:"account",
	FuncName:"updateprofile",
	Input:&backend.ReqUserUpdateProfile{},
	Output:&backend.AckUserUpdateProfile{},
}

var ApiDocReadProfile = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterUser,apibackend.HttpRouterAdmin},
	VerName:"v1",
	SrvName:"account",
	FuncName:"readprofile",
	Input:&backend.ReqUserReadProfile{},
	Output:&backend.AckUserReadProfile{},
}

var ApiDocListUsers = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterAdmin},
	VerName:"v1",
	SrvName:"account",
	FuncName:"listusers",
	Input:&backend.ReqUserList{},
	Output:&backend.AckUserList{Data:[]backend.UserBasic{backend.UserBasic{}}},
}