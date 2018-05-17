package v1

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/api/v1"
	"api_router/base/data"
)

var ApiDocRegister = apidoc.ApiDoc{
	Name:"注册账户",
	Description:"注册钱包用户,管理员权限api",
	Level:data.APILevel_genesis,
	VerName:"v1",
	SrvName:"account",
	FuncName:"register",
	Input:&v1.ReqUserRegister{},
	Output:&v1.AckUserRegister{},
}

var ApiDocUpdateProfile = apidoc.ApiDoc{
	Name:"更新开发设置",
	Description:"更新钱包用户开发信息,管理员权限api",
	Level:data.APILevel_admin,
	VerName:"v1",
	SrvName:"account",
	FuncName:"updateprofile",
	Input:&v1.ReqUserUpdateProfile{},
	Output:&v1.AckUserUpdateProfile{},
}

var ApiDocReadProfile = apidoc.ApiDoc{
	Name:"获取开发设置",
	Description:"获取钱包用户开发信息,管理员权限api",
	Level:data.APILevel_admin,
	VerName:"v1",
	SrvName:"account",
	FuncName:"readprofile",
	Input:&v1.ReqUserReadProfile{},
	Output:&v1.AckUserReadProfile{},
}

var ApiDocListUsers = apidoc.ApiDoc{
	Name:"获取账户信息",
	Description:"获取钱包用户列表,管理员权限api",
	Level:data.APILevel_admin,
	VerName:"v1",
	SrvName:"account",
	FuncName:"listusers",
	Input:&v1.ReqUserList{},
	Output:&v1.AckUserList{Data:[]v1.UserBasic{v1.UserBasic{}}},
}