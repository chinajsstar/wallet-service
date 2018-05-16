package v1

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/api/v1"
	"api_router/base/data"
)

var ApiDocRegister = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"account",
	FuncName:"register",
	Level:data.APILevel_genesis,
	Comment:"注册钱包用户,管理员权限APi",
	Path:"/v1/account/register",
	Input:&v1.ReqUserRegister{},
	Output:&v1.AckUserRegister{},
}

var ApiDocUpdateProfile = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"account",
	FuncName:"updateprofile",
	Level:data.APILevel_admin,
	Comment:"更新钱包用户开发信息,管理员权限APi",
	Path:"/v1/account/updateprofile",
	Input:&v1.ReqUserUpdateProfile{},
	Output:&v1.AckUserUpdateProfile{},
}

var ApiDocReadProfile = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"account",
	FuncName:"readprofile",
	Level:data.APILevel_admin,
	Comment:"获取钱包用户开发信息,管理员权限APi",
	Path:"/v1/account/readprofile",
	Input:&v1.ReqUserReadProfile{},
	Output:&v1.AckUserReadProfile{},
}

var ApiDocListUsers = apidoc.ApiDoc{
	VerName:"v1",
	SrvName:"account",
	FuncName:"listusers",
	Level:data.APILevel_admin,
	Comment:"获取钱包用户列表,管理员权限APi",
	Path:"/v1/account/listusers",
	Input:&v1.ReqUserList{Id:-1},
	Output:&v1.AckUserList{v1.UserBasic{}},
}