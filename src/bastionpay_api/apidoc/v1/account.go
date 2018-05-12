package v1

import (
	"bastionpay_api/apidoc"
	"bastionpay_api/api/v1"
)

var ApiDocRegister = apidoc.ApiDoc{
	Name:"注册钱包用户",
	Comment:"注册钱包用户,管理员权限APi",
	Path:"/api/v1/account/register",
	Input:v1.ReqUserRegister{},
	Output:v1.AckUserRegister{},
}

var ApiDocUpdateProfile = apidoc.ApiDoc{
	Name:"更新钱包用户开发信息",
	Comment:"更新钱包用户开发信息,管理员权限APi",
	Path:"/api/v1/account/updateprofile",
	Input:v1.ReqUserUpdateProfile{},
	Output:v1.AckUserUpdateProfile{},
}