package backend

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/apibackend"
	"bastionpay_api/apibackend/v1/backend"
)

var ApiDocRegister = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterUser,apibackend.HttpRouterAdmin},
	Name:"注册账户",
	Description:"注册钱包用户,管理员权限api",
	VerName:"v1",
	SrvName:"account",
	FuncName:"register",
	Input:&backend.ReqUserRegister{},
	Output:&backend.AckUserRegister{},
}

var ApiDocUpdateProfile = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterUser,apibackend.HttpRouterAdmin},
	Name:"更新开发设置",
	Description:"更新钱包用户开发信息,管理员权限api",
	VerName:"v1",
	SrvName:"account",
	FuncName:"updateprofile",
	Input:&backend.ReqUserUpdateProfile{},
	Output:&backend.AckUserUpdateProfile{},
}

var ApiDocReadProfile = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterUser,apibackend.HttpRouterAdmin},
	Name:"获取开发设置",
	Description:"获取钱包用户开发信息,管理员权限api",
	VerName:"v1",
	SrvName:"account",
	FuncName:"readprofile",
	Input:&backend.ReqUserReadProfile{},
	Output:&backend.AckUserReadProfile{},
}

var ApiDocListUsers = apidoc.ApiDoc{
	Group:[]string{apibackend.HttpRouterAdmin},
	Name:"获取用户列表",
	Description:"获取钱包用户列表,管理员权限api",
	VerName:"v1",
	SrvName:"account",
	FuncName:"listusers",
	Input:&backend.ReqUserList{},
	Output:&backend.AckUserList{Data:[]backend.UserBasic{backend.UserBasic{}}},
}